package service

import (
	"fmt"
	"regexp"
	"strings"
)

var managerDataLeakPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(revenue|sales|profit|earnings|income|target|mtd|ytd)\b.{0,50}(\bTHB\b|[0-9])`),
	regexp.MustCompile(`(?i)(\bTHB\b|[0-9]).{0,50}\b(revenue|sales|profit|earnings|income|target|mtd|ytd)\b`),
	regexp.MustCompile(`(?i)\b(payment mix|payment share|category sales|top products|best[- ]selling)\b.{0,50}([0-9]|%)`),
	regexp.MustCompile(`(ยอดขาย|รายได้|กำไร|เป้า|สินค้าขายดี).{0,50}([0-9]|บาท|%)`),
	regexp.MustCompile(`([0-9][0-9,]*(\.[0-9]+)?\s*(บาท|%)).{0,50}(ยอดขาย|รายได้|กำไร|เป้า|สินค้าขายดี)`),
}

func validateAIResponseGuardrails(access RoleAccess, decision QuestionAuthorizationDecision, reply string) error {
	if access.CanViewManagerData {
		return nil
	}
	if decision.RequiresManagerData {
		return ErrUnsafeAIResponse
	}

	normalizedReply := strings.TrimSpace(reply)
	if normalizedReply == "" {
		return nil
	}
	for _, pattern := range managerDataLeakPatterns {
		if pattern.MatchString(normalizedReply) {
			return fmt.Errorf("%w: response matched restricted pattern %q", ErrUnsafeAIResponse, pattern.String())
		}
	}

	return nil
}
