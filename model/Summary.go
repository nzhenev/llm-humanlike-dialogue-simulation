package model

import (
	"strings"
)

type Summary struct {
	CoreDiscussion    string   `json:"core_discussion"`
	ConfirmedNeeds    []string `json:"confirmed_needs"`
	Constraints       []string `json:"constraints"`
	ExcludedOptions   []string `json:"excluded_options"`
	KeyData           []string `json:"key_data"`
	CurrentConclusion []string `json:"current_conclusion"`
	PendingQuestions  []string `json:"pending_questions"`
	PendingDiscussion []string `json:"pending_discussion"`
}

func (s *Summary) FormatContent() string {
	var builder strings.Builder

	builder.WriteString("[yellow]Core[white]\n")
	builder.WriteString(s.CoreDiscussion + "\n\n")

	addToContent(&builder, "Needs", s.ConfirmedNeeds)
	addToContent(&builder, "Constraints", s.Constraints)
	addToContent(&builder, "Exclude", s.ExcludedOptions)
	addToContent(&builder, "Key Data", s.KeyData)
	addToContent(&builder, "Current", s.CurrentConclusion)
	addToContent(&builder, "Pending Questions", s.PendingQuestions)
	addToContent(&builder, "Pending Discussions", s.PendingDiscussion)

	return builder.String()
}

func addToContent(builder *strings.Builder, title string, list []string) {
	builder.WriteString("[yellow]" + title + "[white]\n")
	for _, item := range list {
		builder.WriteString("• " + item + "\n")
	}
	builder.WriteString("\n")
}

func (s *Summary) FormatContext() string {
	var builder strings.Builder

	builder.WriteString("=== 對話概要 ===\n")
	builder.WriteString("核心討論: " + s.CoreDiscussion + "\n")

	addToContext(&builder, "確認需求", s.ConfirmedNeeds)
	addToContext(&builder, "約束條件", s.Constraints)
	addToContext(&builder, "排除項目", s.ExcludedOptions)
	addToContext(&builder, "關鍵資料", s.KeyData)
	addToContext(&builder, "最新結論", s.CurrentConclusion)
	addToContext(&builder, "待釐清項目", s.PendingQuestions)
	addToContext(&builder, "過往討論", s.PendingDiscussion)

	return builder.String()
}

func addToContext(builder *strings.Builder, title string, list []string) {
	if len(list) > 0 {
		builder.WriteString(title + ":\n")
		for _, question := range list {
			builder.WriteString("- " + question + "\n")
		}
	}
}
