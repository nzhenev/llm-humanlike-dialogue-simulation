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
	appState := model.CreateUI()

	// 設置輸入處理
	appState.Input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			userInput := appState.Input.GetText()
			appState.Input.SetText("")
			appState.APIHandler(userInput)
		}
	})

	// 設置全局按鍵處理
	appState.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			appState.App.Stop()
		case tcell.KeyTab:
			// Tab 切換焦點
			currentFocus := appState.App.GetFocus()
			if currentFocus == appState.Input {
				appState.App.SetFocus(appState.Conversation)
			} else if currentFocus == appState.Conversation {
				appState.App.SetFocus(appState.Summary)
			} else {
				appState.App.SetFocus(appState.Input)
			}
		}
		return event
	})

	// 運行應用
	if err := appState.App.Run(); err != nil {
		panic(err)
	}
}
