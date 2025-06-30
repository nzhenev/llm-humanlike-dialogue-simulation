package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"

	"llmShortTermMemory/model"
)

func init() {
	model.ApiKey = os.Getenv("OPENAI_API_KEY")
	if model.ApiKey == "" {
		data, err := os.ReadFile("OPENAI_API_KEY")
		if err == nil {
			model.ApiKey = strings.TrimSpace(string(data))
		} else {
			execPath, _ := os.Executable()
			execDir := filepath.Dir(execPath)
			configPath := filepath.Join(execDir, "OPENAI_API_KEY")

			data, err := os.ReadFile(configPath)
			if err == nil {
				model.ApiKey = strings.TrimSpace(string(data))
			}
		}
	}

	data, err := os.ReadFile("INSTRUCTION_CONVERSATION")
	if err == nil {
		model.InstructionConversation = strings.TrimSpace(string(data))
	} else {
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		instructionPath := filepath.Join(execDir, "INSTRUCTION_CONVERSATION")

		data, err := os.ReadFile(instructionPath)
		if err == nil {
			model.InstructionConversation = strings.TrimSpace(string(data))
		} else {
			model.InstructionConversation = ""
		}
	}

	data, err = os.ReadFile("INSTRUCTION_SUMMARY")
	if err == nil {
		model.InstructionSummary = strings.TrimSpace(string(data))
	} else {
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		instructionPath := filepath.Join(execDir, "INSTRUCTION_SUMMARY")

		data, err := os.ReadFile(instructionPath)
		if err == nil {
			model.InstructionSummary = strings.TrimSpace(string(data))
		} else {
			model.InstructionSummary = ""
		}
	}
}

func main() {
	var appState *model.Frame

	// 檢查命令列參數
	useOldUI := false
	for _, arg := range os.Args[1:] {
		if arg == "--old" {
			useOldUI = true
			break
		}
	}

	if useOldUI {
		appState = model.CreateOldUI()
		appState.Input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			text := appState.Input.GetText()
			if event.Key() == tcell.KeyEnter && len(text) > 2 && strings.HasSuffix(text, "$$") {
				text = strings.TrimSuffix(text, "$$")
				appState.Input.SetText("", true)
				appState.OldAPIHandler(text)
				return nil
			}
			return event
		})

	} else {
		appState = model.CreateUI()
		appState.Input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			text := appState.Input.GetText()
			if event.Key() == tcell.KeyEnter && len(text) > 2 && strings.HasSuffix(text, "$$") {
				text = strings.TrimSuffix(text, "$$")
				appState.Input.SetText("", true)
				appState.APIHandler(text)
				return nil
			}
			return event
		})
	}

	// 設置全局按鍵處理
	appState.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			appState.App.Stop()
		case tcell.KeyTab:
			// Tab 切換焦點
			currentFocus := appState.App.GetFocus()
			if useOldUI {
				// 舊版 UI 只有 Input 和 Conversation
				if currentFocus == appState.Input {
					appState.App.SetFocus(appState.Conversation)
				} else {
					appState.App.SetFocus(appState.Input)
				}
			} else {
				// 新版 UI 有 Input、Conversation 和 Summary
				if currentFocus == appState.Input {
					appState.App.SetFocus(appState.Conversation)
				} else if currentFocus == appState.Conversation {
					appState.App.SetFocus(appState.Summary)
				} else {
					appState.App.SetFocus(appState.Input)
				}
			}
		}
		return event
	})

	// 運行應用
	if err := appState.App.Run(); err != nil {
		panic(err)
	}
}
