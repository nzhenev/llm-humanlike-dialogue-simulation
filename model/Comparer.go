// TODO: 主要依賴網路的寫法，此頁算法需額外加強
package model

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type ConversationRecord struct {
	ID      int       `json:"id"`
	SendAt  time.Time `json:"send_at"`
	User    string    `json:"user"`
	Content string    `json:"content"`
	Keyword []string  `json:"keyword"`
}

type SearchResult struct {
	Record *ConversationRecord `json:"record"`
	Score  float64             `json:"score"`
}

type Comparer struct {
	recordList []*ConversationRecord
	threshold  float64
}

func NewFuzzyComparer(threshold float64) *Comparer {
	return &Comparer{
		recordList: make([]*ConversationRecord, 0),
		threshold:  threshold,
	}
}

func (f *Comparer) AddRecord(speaker, content string) {
	record := &ConversationRecord{
		ID:      len(f.recordList) + 1,
		SendAt:  time.Now(),
		User:    speaker,
		Content: content,
		Keyword: getKeywordList(content),
	}
	f.recordList = append(f.recordList, record)
}

func (f *Comparer) Search(query string) []*ConversationRecord {
	if len(f.recordList) == 0 {
		return nil
	}

	keywordList := getKeywordList(query)
	resultList := make([]SearchResult, 0)

	// 對每個歷史記錄計算相關性分數
	for _, record := range f.recordList {
		score := f.calcScore(keywordList, record, query)
		if score >= f.threshold {
			resultList = append(resultList, SearchResult{
				Record: record,
				Score:  score,
			})
		}
	}

	// 按分數排序
	sort.Slice(resultList, func(i, j int) bool {
		return resultList[i].Score > resultList[j].Score
	})

	// 相關記錄
	relevantRecords := make([]*ConversationRecord, 0, len(resultList))
	for _, result := range resultList {
		relevantRecords = append(relevantRecords, result.Record)
	}

	return relevantRecords
}

func (f *Comparer) calcScore(keywordList []string, record *ConversationRecord, query string) float64 {
	keyword := f.calcKeyword(keywordList, record.Keyword)
	semantic := f.calcSemantic(query, record.Content)
	time := f.calcTime(record.SendAt)

	return keyword*0.4 + semantic*0.4 + time*0.2
}

// 計算關鍵詞重疊
func (f *Comparer) calcKeyword(queryKeywordList, recordKeywordList []string) float64 {
	if len(queryKeywordList) == 0 || len(recordKeywordList) == 0 {
		return 0.0
	}

	matches := 0
	for _, qk := range queryKeywordList {
		for _, rk := range recordKeywordList {
			if strings.Contains(strings.ToLower(qk), strings.ToLower(rk)) ||
				strings.Contains(strings.ToLower(rk), strings.ToLower(qk)) {
				matches++
				break
			}
		}
	}

	union := len(queryKeywordList) + len(recordKeywordList) - matches
	if union == 0 {
		return 0.0
	}

	return float64(matches) / float64(union)
}

// 計算語義相似度
func (f *Comparer) calcSemantic(query, content string) float64 {
	queryWordList := strings.Fields(strings.ToLower(query))
	contentWordList := strings.Fields(strings.ToLower(content))

	if len(queryWordList) == 0 || len(contentWordList) == 0 {
		return 0.0
	}

	// 計算共同詞彙比例
	count := 0
	for _, qw := range queryWordList {
		for _, cw := range contentWordList {
			if qw == cw {
				count++
				break
			}
		}
	}

	// 使用餘弦相似度概念
	return float64(count) / math.Sqrt(float64(len(queryWordList)*len(contentWordList)))
}

// 計算時間衰減分數
func (f *Comparer) calcTime(timestamp time.Time) float64 {
	now := time.Now()
	duration := now.Sub(timestamp)
	hours := duration.Hours()

	// 24小時內線性衰減：最近=1.0，24小時前=0.7
	// 超過24小時後維持固定分數0.7
	if hours <= 24 {
		return 1.0 - (hours * 0.3 / 24.0)
	}

	// 超過24小時後固定分數
	return 0.7
}

// TODO: 需研究提取關鍵詞的寫法
// 提取關鍵詞
func getKeywordList(text string) []string {
	// 停用詞列表
	stopList := map[string]bool{
		"的": true, "是": true, "在": true, "有": true, "和": true,
		"與": true, "或": true, "但": true, "這": true, "那": true,
		"我": true, "你": true, "他": true, "她": true, "它": true,
		"了": true, "嗎": true, "呢": true, "啊": true, "吧": true,
		"the": true, "is": true, "at": true, "which": true, "on": true,
		"and": true, "or": true, "but": true, "this": true, "that": true,
		"i": true, "you": true, "he": true, "she": true, "it": true,
	}

	wordList := strings.Fields(strings.ToLower(text))
	keywordList := make([]string, 0)

	for _, word := range wordList {
		// 移除標點符號
		word = strings.Trim(word, ".,!?;:()[]{}\"'")

		// 過濾停用詞和短詞
		if len(word) >= 2 && !stopList[word] {
			keywordList = append(keywordList, word)
		}
	}

	return keywordList
}

// 格式化相關上下文
func (f *Comparer) FormatRelevant(records []*ConversationRecord) string {
	if len(records) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("=== 相關歷史對話 ===\n")

	for i, record := range records {
		if i >= 5 {
			break
		}

		speakerName := "User"
		if record.User == "assistant" {
			speakerName = "LLM"
		}

		builder.WriteString(fmt.Sprintf("%s: %s\n", speakerName, record.Content))
	}

	builder.WriteString("\n")
	return builder.String()
}
