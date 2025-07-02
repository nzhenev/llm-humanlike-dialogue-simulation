package model

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pkoukk/tiktoken-go"
	"github.com/rivo/tview"
)

type Frame struct {
	App             *tview.Application
	Conversation    *tview.TextView
	Summary         *tview.TextView
	Input           *tview.TextArea
	CurrentSummary  Summary
	conversationLog strings.Builder
	Comparer        *Comparer
}

func CreateUI() *Frame {
	app := tview.NewApplication()

	conversationView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)
	conversationView.
		SetBorder(true).
		SetTitle(" Record ").
		SetTitleAlign(tview.AlignLeft)

	summaryView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)
	summaryView.
		SetBorder(true).
		SetTitle(" Summary ").
		SetTitleAlign(tview.AlignLeft)

	inputField := tview.NewTextArea().
		SetLabel("Input: ").
		SetWrap(true).
		SetWordWrap(true)
	inputField.
		SetBorder(true).
		SetTitle(" Message ").
		SetTitleAlign(tview.AlignLeft)

	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(summaryView, 0, 2, true).
		AddItem(inputField, 9, 0, true)

	mainFlex := tview.NewFlex().
		AddItem(conversationView, 0, 2, true).
		AddItem(rightFlex, 0, 1, true)

	app.SetRoot(mainFlex, true).SetFocus(inputField)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			current := app.GetFocus()
			if current == conversationView {
				app.SetFocus(summaryView)
			} else if current == summaryView {
				app.SetFocus(inputField)
			} else {
				app.SetFocus(conversationView)
			}
			return nil
		}
		return event
	})

	summary := Summary{
		CoreDiscussion:    "empty",
		ConfirmedNeeds:    []string{},
		Constraints:       []string{},
		ExcludedOptions:   []string{},
		KeyData:           []string{},
		CurrentConclusion: []string{},
		PendingQuestions:  []string{},
		PendingDiscussion: []string{},
	}

	fuzzySearcher := NewFuzzyComparer(0.3)

	frame := &Frame{
		Conversation:   conversationView,
		Summary:        summaryView,
		Input:          inputField,
		App:            app,
		CurrentSummary: summary,
		Comparer:       fuzzySearcher,
	}

	summaryView.SetText(summary.FormatContent())

	now := time.Now().Format("15:04:05")
	msg := fmt.Sprintf("[gray]%s[white] [green]LLM[white]: Type to start chat\n[yellow]Shortcuts[white]: Type message and end with $$ to send | Tab to Switch Panel | Ctrl+C to Exit\n\n", now)
	frame.conversationLog.WriteString(msg)
	conversationView.SetText(frame.conversationLog.String())

	return frame
}

func (f *Frame) AddToConversation(setTime bool, speaker, message string) {
	now := time.Now().Format("15:04:05")
	var msg string

	if setTime {
		msg = fmt.Sprintf("[gray]%s[white] %s: %s\n\n", now, speaker, message)
	} else {
		msg = fmt.Sprintf("%s: %s\n\n", speaker, message)
	}

	f.conversationLog.WriteString(msg)
	f.Conversation.SetText(f.conversationLog.String())
	f.Conversation.ScrollToEnd()

	if f.Comparer != nil {
		speakerType := "user"
		if strings.Contains(speaker, "LLM") {
			speakerType = "assistant"
		}
		f.Comparer.AddRecord(speakerType, message)
	}
}

func (f *Frame) updateSummary() {
	f.Summary.SetText(f.CurrentSummary.FormatContent())
}

func (f *Frame) APIHandler(userInput string) {
	if userInput == "" {
		return
	}

	userInput = strings.TrimSpace(userInput)

	systemPrompt := strings.TrimSpace(
		fmt.Sprintf(`=== 系統資訊 ===
當前時間：%s
作業系統與執行環境：%s

=== 指令說明 ===
%s`,
			time.Now().Format("2006年01月02日 15:04:05"),
			runtime.GOOS+"/"+runtime.GOARCH,
			InstructionConversation,
		),
	)
	systemSummary := strings.TrimSpace(f.CurrentSummary.FormatContext())

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "system",
			Content: systemSummary,
		},
	}

	f.AddToConversation(true, fmt.Sprintf("[yellow]%v[white]", "User"), userInput)

	// 使用模糊搜尋找到相關歷史對話
	relevantRecords := f.Comparer.Search(userInput)
	relevantContext := strings.TrimSpace(f.Comparer.FormatRelevant(relevantRecords))

	// 構建包含相關歷史的上下文
	if relevantContext != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: relevantContext,
		})
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: userInput,
	})

	// 使用 tiktoken 來計算 token 數量 for gpt-4o
	tke, err := tiktoken.GetEncoding("o200k_base")
	if err != nil {
		f.AddToConversation(true, fmt.Sprintf("[red]%v[white]", "錯誤"), fmt.Sprintf("[red]%v[white]", err))
		return
	}

	systemToken := tke.Encode(systemPrompt+systemSummary+relevantContext, nil, nil)
	inputToken := tke.Encode(userInput, nil, nil)

	f.AddToConversation(false, fmt.Sprintf("[grey]%v[white]", "Request token"), fmt.Sprintf("[grey]%d[white]", len(inputToken)+len(systemToken)))

	go func() {
		response, err := askWithLargeModel(messages)

		f.App.QueueUpdateDraw(func() {
			if err != nil {
				f.AddToConversation(true, fmt.Sprintf("[red]%v[white]", "Error"), fmt.Sprintf("[red]%v[white]", err))
				return
			}

			f.AddToConversation(true, fmt.Sprintf("[green]%v[white]", "LLM"), response)

			go func() {
				newSummary := f.generateSummary(f.CurrentSummary, userInput, response)
				f.App.QueueUpdateDraw(func() {
					f.CurrentSummary = newSummary
					f.updateSummary()
				})
			}()
		})
	}()
}

func (f *Frame) generateSummary(summary Summary, input, assistant string) Summary {
	prompt := fmt.Sprintf(`基於以下資訊更新對話概要，保持 JSON 格式：

當前概要：
%s

新對話：
用戶問題：%s
助手回覆：%s

%s

JSON 格式要求：
{
  "core_discussion": "當前討論的核心主題",
  "confirmed_needs": ["累積保留所有確認的需求"],
  "constraints": ["累積保留所有約束條件"],
  "excluded_options": ["被排除的選項：原因（敏感識別用戶排除意圖）"],
  "key_data": ["累積保留所有重要資料和事實"],
  "current_conclusion": ["按時間順序的所有結論，最新在前"],
  "pending_questions": ["當前主題相關的待釐清問題"],
  "pending_discussion": ["所有重要的歷史討論點，包括之前的主題"]
}

只回傳 JSON，不要其他說明。

請更新概要：`,
		summary.FormatContext(),
		input,
		assistant,
		InstructionSummary,
	)

	messages := []Message{
		{
			Role:    "system",
			Content: "你是一個專業的對話概要整理助手。請根據對話內容提取並更新概要，保持 JSON 格式輸出。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// 使用 tiktoken 來計算 token 數量 for gpt-4o
	tke, err := tiktoken.GetEncoding("o200k_base")
	if err != nil {
		f.AddToConversation(true, fmt.Sprintf("[red]%v[white]", "錯誤"), fmt.Sprintf("[red]%v[white]", err))
		return summary
	}

	promptToken := tke.Encode(prompt, nil, nil)
	systemToken := tke.Encode("你是一個專業的對話概要整理助手。請根據對話內容提取並更新概要，保持 JSON 格式輸出。", nil, nil)

	f.AddToConversation(false, fmt.Sprintf("[grey]%v[white]", "Summary token"), fmt.Sprintf("[grey]%d[white]", len(promptToken)+len(systemToken)))

	response, err := askWithSmallModel(messages)
	if err != nil {
		f.AddToConversation(true, fmt.Sprintf("[red]%v[white]", "錯誤"), fmt.Sprintf("[red]%v[white]", err))
		return summary
	}

	result := strings.TrimSpace(response)
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	var newSummary Summary
	if err := json.Unmarshal([]byte(result), &newSummary); err != nil {
		return summary
	}

	return newSummary
}

func CreateOldUI() *Frame {
	app := tview.NewApplication()

	conversationView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)
	conversationView.
		SetBorder(true).
		SetTitle(" Record ").
		SetTitleAlign(tview.AlignLeft)

	inputField := tview.NewTextArea().
		SetLabel("Input: ").
		SetWrap(true).
		SetWordWrap(true)
	inputField.
		SetBorder(true).
		SetTitle(" Message ").
		SetTitleAlign(tview.AlignLeft)

	contentFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(conversationView, 0, 2, true).
		AddItem(inputField, 9, 0, true)

	mainFlex := tview.NewFlex().
		AddItem(contentFlex, 0, 1, true)

	app.SetRoot(mainFlex, true).SetFocus(inputField)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			current := app.GetFocus()
			if current == conversationView {
				app.SetFocus(inputField)
			} else {
				app.SetFocus(conversationView)
			}
			return nil
		}
		return event
	})

	frame := &Frame{
		Conversation: conversationView,
		Input:        inputField,
		App:          app,
	}

	now := time.Now().Format("15:04:05")
	msg := fmt.Sprintf("[gray]%s[white] [green]LLM[white]: Type to start chat\n[yellow]Shortcuts[white]: Type message and end with $$ to send | Tab to Switch Panel | Ctrl+C to Exit\n\n", now)
	frame.conversationLog.WriteString(msg)
	conversationView.SetText(frame.conversationLog.String())

	return frame
}

func (f *Frame) OldAPIHandler(userInput string) {
	if userInput == "" {
		return
	}

	userInput = strings.TrimSpace(userInput)

	f.AddToConversation(true, fmt.Sprintf("[yellow]%v[white]", "User"), userInput)

	systemPrompt := strings.TrimSpace(
		fmt.Sprintf(`=== 系統資訊 ===
當前時間：%s
作業系統與執行環境：%s

=== 指令說明 ===
%s`,
			time.Now().Format("2006年01月02日 15:04:05"),
			runtime.GOOS+"/"+runtime.GOARCH,
			InstructionConversation,
		),
	)

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	conversationText := f.conversationLog.String()
	lines := strings.Split(conversationText, "\n")

	for _, line := range lines {
		// 修正解析邏輯，處理顏色標記
		if strings.Contains(line, "[yellow]User[white]:") {
			parts := strings.SplitN(line, "[yellow]User[white]: ", 2)
			if len(parts) > 1 {
				messages = append(messages, Message{
					Role:    "user",
					Content: parts[1],
				})
			}
		} else if strings.Contains(line, "[green]LLM[white]:") {
			parts := strings.SplitN(line, "[green]LLM[white]: ", 2)
			if len(parts) > 1 {
				messages = append(messages, Message{
					Role:    "assistant",
					Content: parts[1],
				})
			}
		}
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: userInput,
	})

	tke, err := tiktoken.GetEncoding("o200k_base")
	if err != nil {
		f.AddToConversation(true, fmt.Sprintf("[red]%v[white]", "錯誤"), fmt.Sprintf("[red]%v[white]", err))
		return
	}

	totalTokens := 0
	for _, msg := range messages {
		tokens := tke.Encode(msg.Content, nil, nil)
		totalTokens += len(tokens)
	}

	f.AddToConversation(false, fmt.Sprintf("[grey]%v[white]", "Request token"), fmt.Sprintf("[grey]%d[white]", totalTokens))

	go func() {
		response, err := askWithLargeModel(messages)

		f.App.QueueUpdateDraw(func() {
			if err != nil {
				f.AddToConversation(true, fmt.Sprintf("[red]%v[white]", "Error"), fmt.Sprintf("[red]%v[white]", err))
				return
			}

			f.AddToConversation(true, fmt.Sprintf("[green]%v[white]", "LLM"), response)
		})
	}()
}
