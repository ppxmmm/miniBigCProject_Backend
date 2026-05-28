package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/genai"
)

// ErrUnclearQuestionAuthorization is returned when the classifier cannot make a safe decision.
var ErrUnclearQuestionAuthorization = errors.New("question authorization classification is unclear")

// QuestionAuthorizationDecision is the semantic classification used for role authorization.
type QuestionAuthorizationDecision struct {
	RequiresManagerData bool     `json:"requires_manager_data"`
	Domains             []string `json:"domains"`
	Confidence          float64  `json:"confidence"`
	Reason              string   `json:"reason"`
	Source              string   `json:"source,omitempty"`
}

// QuestionAuthorizationService classifies whether a question needs manager-only data.
type QuestionAuthorizationService interface {
	ClassifyQuestion(ctx context.Context, question string, access RoleAccess) (QuestionAuthorizationDecision, error)
}

type compositeQuestionAuthorizationService struct {
	deterministic QuestionAuthorizationService
	fallback      QuestionAuthorizationService
}

// NewQuestionAuthorizationService creates the production guardrail classifier chain.
func NewQuestionAuthorizationService(apiKey string, model string, timeout time.Duration) QuestionAuthorizationService {
	return &compositeQuestionAuthorizationService{
		deterministic: NewDeterministicQuestionAuthorizationService(),
		fallback:      NewGeminiQuestionAuthorizationService(apiKey, model, timeout),
	}
}

func (service *compositeQuestionAuthorizationService) ClassifyQuestion(ctx context.Context, question string, access RoleAccess) (QuestionAuthorizationDecision, error) {
	decision, err := service.deterministic.ClassifyQuestion(ctx, question, access)
	if err == nil {
		return decision, nil
	}
	if !errors.Is(err, ErrUnclearQuestionAuthorization) {
		return QuestionAuthorizationDecision{}, err
	}

	log.Printf(
		"AI question authorization classifier handoff: role=%s store_id=%d from=semantic_deterministic to=gemini_classifier reason=unclear",
		access.Role,
		access.DashboardStoreID,
	)
	decision, err = service.fallback.ClassifyQuestion(ctx, question, access)
	if errors.Is(err, ErrMissingGeminiAPIKey) {
		return QuestionAuthorizationDecision{}, ErrUnclearQuestionAuthorization
	}
	return decision, err
}

type deterministicQuestionAuthorizationService struct{}

type domainScore struct {
	domain              string
	requiresManagerData bool
	score               int
}

type domainRule struct {
	domain              string
	requiresManagerData bool
	weight              int
	phrases             []string
}

// NewDeterministicQuestionAuthorizationService creates the local semantic classifier.
func NewDeterministicQuestionAuthorizationService() QuestionAuthorizationService {
	return deterministicQuestionAuthorizationService{}
}

func (service deterministicQuestionAuthorizationService) ClassifyQuestion(_ context.Context, question string, access RoleAccess) (QuestionAuthorizationDecision, error) {
	if access.CanViewManagerData {
		return QuestionAuthorizationDecision{
			RequiresManagerData: false,
			Domains:             []string{"manager_access"},
			Confidence:          1,
			Reason:              "Role can view manager data.",
			Source:              "semantic_deterministic",
		}, nil
	}

	normalized := normalizeQuestionText(question)
	if normalized == "" {
		return QuestionAuthorizationDecision{}, ErrUnclearQuestionAuthorization
	}

	scores := scoreQuestionDomains(normalized)
	winner, ok := highestQuestionDomainScore(scores)
	if !ok || winner.score < 3 {
		return QuestionAuthorizationDecision{}, ErrUnclearQuestionAuthorization
	}

	confidence := deterministicConfidence(winner.score)
	if confidence < 0.70 {
		return QuestionAuthorizationDecision{}, ErrUnclearQuestionAuthorization
	}

	return QuestionAuthorizationDecision{
		RequiresManagerData: winner.requiresManagerData,
		Domains:             []string{winner.domain},
		Confidence:          confidence,
		Reason:              fmt.Sprintf("Deterministic domain score classified the question as %s.", winner.domain),
		Source:              "semantic_deterministic",
	}, nil
}

type geminiQuestionAuthorizationService struct {
	apiKey  string
	model   string
	timeout time.Duration
}

// NewGeminiQuestionAuthorizationService creates a Gemini-backed semantic authorization classifier.
func NewGeminiQuestionAuthorizationService(apiKey string, model string, timeout time.Duration) QuestionAuthorizationService {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-2.5-flash"
	}
	if timeout <= 0 {
		timeout = defaultAITimeout
	}

	return &geminiQuestionAuthorizationService{
		apiKey:  strings.TrimSpace(apiKey),
		model:   model,
		timeout: timeout,
	}
}

func (service *geminiQuestionAuthorizationService) ClassifyQuestion(ctx context.Context, question string, access RoleAccess) (QuestionAuthorizationDecision, error) {
	if access.CanViewManagerData {
		return QuestionAuthorizationDecision{
			RequiresManagerData: false,
			Domains:             []string{"manager_access"},
			Confidence:          1,
			Reason:              "Role can view manager data.",
			Source:              "gemini_classifier",
		}, nil
	}
	if service.apiKey == "" {
		return QuestionAuthorizationDecision{}, ErrMissingGeminiAPIKey
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), service.timeout)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  service.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return QuestionAuthorizationDecision{}, fmt.Errorf("create Gemini authorization client: %w", err)
	}

	prompt := fmt.Sprintf(semanticQuestionAuthorizationPrompt, strings.TrimSpace(access.Role), strings.TrimSpace(question))
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: semanticQuestionAuthorizationSystemInstruction}},
		},
		Temperature: genai.Ptr[float32](0),
	}

	response, err := client.Models.GenerateContent(ctx, service.model, genai.Text(prompt), config)
	if err != nil {
		return QuestionAuthorizationDecision{}, wrapGeminiError("classify question authorization", err)
	}

	text := collectResponseText(response)
	decision, err := parseQuestionAuthorizationDecision(text)
	if err != nil {
		return QuestionAuthorizationDecision{}, err
	}
	if decision.Confidence < 0.60 {
		return QuestionAuthorizationDecision{}, ErrUnclearQuestionAuthorization
	}
	decision.Source = "gemini_classifier"

	return decision, nil
}

func parseQuestionAuthorizationDecision(text string) (QuestionAuthorizationDecision, error) {
	payload := extractJSONObject(text)
	if payload == "" {
		return QuestionAuthorizationDecision{}, ErrUnclearQuestionAuthorization
	}

	var decision QuestionAuthorizationDecision
	if err := json.Unmarshal([]byte(payload), &decision); err != nil {
		return QuestionAuthorizationDecision{}, fmt.Errorf("%w: %v", ErrUnclearQuestionAuthorization, err)
	}
	decision.Reason = strings.TrimSpace(decision.Reason)
	for i, domain := range decision.Domains {
		decision.Domains[i] = strings.TrimSpace(domain)
	}

	return decision, nil
}

func normalizeQuestionText(question string) string {
	normalized := strings.ToLower(strings.TrimSpace(question))
	replacer := strings.NewReplacer(
		"?", " ",
		"!", " ",
		".", " ",
		",", " ",
		":", " ",
		";", " ",
		"/", " ",
		"-", " ",
		"_", " ",
		"\n", " ",
		"\t", " ",
	)
	normalized = replacer.Replace(normalized)
	return strings.Join(strings.Fields(normalized), " ")
}

func scoreQuestionDomains(question string) map[string]domainScore {
	scores := make(map[string]domainScore)
	for _, rule := range questionDomainRules {
		for _, phrase := range rule.phrases {
			if strings.Contains(question, normalizeQuestionText(phrase)) {
				score := scores[rule.domain]
				score.domain = rule.domain
				score.requiresManagerData = rule.requiresManagerData
				score.score += rule.weight
				scores[rule.domain] = score
			}
		}
	}
	return scores
}

func highestQuestionDomainScore(scores map[string]domainScore) (domainScore, bool) {
	var winner domainScore
	for _, score := range scores {
		if score.requiresManagerData && !winner.requiresManagerData {
			winner = score
			continue
		}
		if score.requiresManagerData == winner.requiresManagerData && score.score > winner.score {
			winner = score
		}
	}
	return winner, winner.score > 0
}

func deterministicConfidence(score int) float64 {
	switch {
	case score >= 8:
		return 0.95
	case score >= 5:
		return 0.85
	case score >= 3:
		return 0.72
	default:
		return 0
	}
}

func extractJSONObject(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return ""
	}
	return text[start : end+1]
}

const semanticQuestionAuthorizationSystemInstruction = `You classify whether a store dashboard question requires manager-only data.
Return only strict JSON. Do not include markdown or extra text.
Manager-only domains include sales performance, revenue, target, MTD/YTD, payment mix, category sales, top products, profit, promotions, suggestions, and business recovery plans.
Staff-allowed domains include store profile, inventory quantity, low stock, expiring inventory, delivery orders, late deliveries, and operational fulfillment status.
If a question is ambiguous and might require manager-only data, set requires_manager_data to true.`

const semanticQuestionAuthorizationPrompt = `Role: %s
Question: %q

Classify the question into one or more business domains and decide whether answering it requires manager-only data.
Respond exactly as JSON:
{
  "requires_manager_data": true,
  "domains": ["sales_performance"],
  "confidence": 0.95,
  "reason": "brief reason"
}`

var questionDomainRules = []domainRule{
	{
		domain:              "sales_performance",
		requiresManagerData: true,
		weight:              4,
		phrases: []string{
			"how much money",
			"money did we make",
			"money we made",
			"made today",
			"revenue",
			"sales",
			"ยอดขาย",
			"รายได้",
			"income",
			"earnings",
			"profit",
			"below target",
			"hit target",
			"against target",
			"mtd",
			"ytd",
			"เป้า",
			"performance",
			"recovery gap",
		},
	},
	{
		domain:              "payment_mix",
		requiresManagerData: true,
		weight:              4,
		phrases: []string{
			"payment mix",
			"payment share",
			"qr payment",
			"cash share",
			"card share",
			"wallet share",
			"checkout behavior",
			"วิธีจ่าย",
			"การชำระเงิน",
		},
	},
	{
		domain:              "top_products",
		requiresManagerData: true,
		weight:              4,
		phrases: []string{
			"top products",
			"best selling",
			"best-selling",
			"most money",
			"made the most",
			"product revenue",
			"hero sku",
			"สินค้าไหนขายดี",
			"สินค้าขายดี",
		},
	},
	{
		domain:              "category_sales",
		requiresManagerData: true,
		weight:              4,
		phrases: []string{
			"category sales",
			"category revenue",
			"category share",
			"which category",
			"หมวดไหนขายดี",
		},
	},
	{
		domain:              "suggestions_promotions",
		requiresManagerData: true,
		weight:              4,
		phrases: []string{
			"suggestion",
			"recommendation",
			"promotion",
			"promo",
			"discount campaign",
			"business recovery",
			"แนะนำ",
			"โปรโมชัน",
			"โปรโมชั่น",
		},
	},
	{
		domain:              "product_availability",
		requiresManagerData: false,
		weight:              2,
		phrases: []string{
			"available",
			"in stock",
			"do we have",
			"have milk",
			"มี",
			"ไหม",
			"มั้ย",
			"มีของ",
			"มีสินค้า",
		},
	},
	{
		domain:              "operational_actions",
		requiresManagerData: false,
		weight:              4,
		phrases: []string{
			"action items",
			"to do",
			"todo",
			"tasks",
			"what should i do",
			"what needs to be done",
			"ต้องทำอะไร",
			"ต้องดำเนินการ",
			"รายการที่ต้องดำเนินการ",
			"งานที่ต้องทำ",
			"มีอะไรต้องทำ",
			"อะไรที่ต้องปรับปรุง",
			"มีอะไรที่ต้องปรับปรุง",
			"วันนี้มีอะไรที่ต้องปรับปรุง",
			"วันนี้ปรับปรุงอะไร",
			"วันนี้ควรปรับปรุงอะไร",
			"วันนี้ต้องปรับปรุงอะไร",
			"วันนี้มีอะไรต้องปรับปรุง",
			"วันนี้ต้องแก้อะไร",
			"วันนี้ควรแก้อะไร",
			"วันนี้มีอะไรต้องแก้",
			"วันนี้ต้องจัดการอะไร",
			"วันนี้ควรจัดการอะไร",
			"วันนี้มีอะไรต้องจัดการ",
			"วันนี้ควรทำอะไร",
			"วันนี้ต้องทำอะไร",
			"วันนี้ควรโฟกัสอะไร",
			"วันนี้ต้องโฟกัสอะไร",
			"วันนี้ควรดูอะไร",
			"วันนี้ต้องดูอะไร",
			"วันนี้มีอะไรน่าห่วง",
			"วันนี้มีอะไรผิดปกติ",
			"วันนี้มีปัญหาอะไร",
			"วันนี้ต้องรีบทำอะไร",
			"วันนี้ควรรีบทำอะไร",
			"ควรปรับปรุงอะไร",
			"ควรปรับปรุงอะไรไหม",
			"ควรปรับปรุงอะไรมั้ย",
			"ต้องปรับปรุงอะไร",
			"ต้องแก้อะไร",
			"ควรแก้อะไร",
			"ต้องจัดการอะไร",
			"ควรจัดการอะไร",
			"ควรโฟกัสอะไร",
			"ต้องโฟกัสอะไร",
			"ควรดูอะไร",
			"ต้องดูอะไร",
			"มีอะไรน่าห่วง",
			"มีอะไรผิดปกติ",
			"มีปัญหาอะไร",
			"ควรรีบทำอะไร",
			"ต้องรีบจัดการ",
		},
	},
	{
		domain:              "low_stock",
		requiresManagerData: false,
		weight:              4,
		phrases: []string{
			"low stock",
			"running out",
			"out of stock",
			"almost out",
			"reorder",
			"restock",
			"เติมของ",
			"ใกล้หมด",
			"ของหมด",
			"สต็อกต่ำ",
		},
	},
	{
		domain:              "expiring_inventory",
		requiresManagerData: false,
		weight:              4,
		phrases: []string{
			"expire",
			"expiry",
			"expiring",
			"near expiry",
			"หมดอายุ",
			"ใกล้หมดอายุ",
		},
	},
	{
		domain:              "deliveries",
		requiresManagerData: false,
		weight:              4,
		phrases: []string{
			"delivery",
			"deliveries",
			"late order",
			"late delivery",
			"otif",
			"order status",
			"ส่งของ",
			"ออเดอร์",
			"ส่งช้า",
		},
	},
	{
		domain:              "store_profile",
		requiresManagerData: false,
		weight:              4,
		phrases: []string{
			"store address",
			"store profile",
			"manager name",
			"branch code",
			"ที่อยู่ร้าน",
			"รหัสสาขา",
		},
	},
}
