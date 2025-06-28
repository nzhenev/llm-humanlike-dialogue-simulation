> [!Note]
> This content is translated by LLM. Original text can be found [here](README.zh.md)

# LLM Short-Term Memory Solutions
> Solving the "LLM gets lost in multi-turn conversations" problem by **human-like dialogue simulation**.<br>
> Reference paper: [LLMs Get Lost In Multi-Turn Conversation](https://arxiv.org/abs/2505.06120)

## Real Conversation Process

### Continuous Summary Updates in Mind
- Humans don't repeat entire conversation history before responding
- Instead, they maintain dynamic "current understanding" and update perspectives based on new information
- Past details fade or disappear, but key conclusions and constraints persist

### Keyword-Triggered Recall
- When someone says "what we discussed earlier"
- We perform fuzzy search in recent memory for relevant information
- Specific details are only retrieved when triggered by reference keywords

### Implementation Plan
```
Human Conversation Process → Engineering Implementation

Mental Summary → Generate structured summaries after each turn using lightweight models
Content Recall → Automatic fuzzy search of conversation history (relevance scoring)
New Conversation → Latest summary + relevant history fragments + new question
```

## Paper Problem Analysis
> [LLMs Get Lost In Multi-Turn Conversation](https://arxiv.org/abs/2505.06120)

### Common LLM Issues in Long Conversations
Research on **15 LLMs across 200,000+ conversations** shows:
- Problem:<br>
  39% performance degradation in multi-turn conversations
- Cause:<br>
  LLMs use "complete memory" model instead of human-like selective memory

### Four Major Issues

| LLM Problem | Human Behavior | Proposed Solution |
| - | - | - |
| **Premature Solutions** | Ask clarifying questions first | - |
| **Information Hoarding** | Forget irrelevant details | Structured summaries retain only relevant info |
| **Linear Replay** | Maintain "current state" | Refresh summary each turn + dynamic history retrieval |
| **Verbose Spiraling** | Stay focused | - |

## Memory Architecture Comparison

### LLM "Complete Memory" (Non-human conversation style)
```
Turn 1: [question 1] + [response 1]
Turn 2: [question 1] + [response 1] + [question 2] + [response 2]
Turn 3: [question 1] + [response 1] + [question 2] + [response 2] + [question 3] + [response 3]
...
Turn N: [Complete verbatim conversation record]
```
- Humans don't perfectly recall all content
- Old irrelevant information interferes with current content generation; humans filter out irrelevant info
- No mechanism to learn from mistakes; gets distracted by irrelevant info in long conversations
- Linear token growth limits conversation length; humans don't interrupt conversations due to length

### Simulating Real Human Conversation Methods
```
Each Turn: [Structured current state] + [relevant history fragments] + [new question]
```

#### Conversation Summary Design
```
Core topic of current discussion
Accumulated confirmed requirements
Accumulated constraint conditions
Excluded options + reasons
Accumulated important data, facts, and conclusions
Current topic-related questions to clarify
All important historical discussion points
```

## Human Conversation Simulation
> **Simulating incomplete memory**: Not designing better, larger information retrieval support, but building systems that process information like humans

### Humans Are Inherently Poor at Complete Memory
- We forget irrelevant details
- We remember key decisions
- We learn from mistakes
- We have internal measuring sticks
- We maintain current conversation focus
- We actively associate relevant past content

## Filtering Beats Memory

**Cognitive Burden of Perfect Memory**

Research shows exceptional memory ability doesn't equal intelligence advantage. The classic neuropsychological case [Solomon Shereshevsky](https://en.wikipedia.org/wiki/Solomon_Shereshevsky) could remember arbitrary details from decades ago, including meaningless number sequences and vocabulary lists, but "perfect memory" actually created cognitive burden. He couldn't distinguish important from unimportant information, leading to difficulties in abstract thinking and daily decision-making.

### Design Insights for LLMs
Traditional LLMs using complete memory models may actually be simulating cognitive disabilities. This leads to requiring increasingly powerful hardware support without proportional performance gains.
```
Selective attention > Complete memory
Abstract summarization > Detail preservation
Dynamic adaptation > Fixed replay
```

### Combining Machine Advantages
This approach explores combining human cognitive advantages with machine computational strengths:
- Simulating human mechanisms: Default state uses only structured summaries, avoiding historical information overwhelming current conversation
- Machine enhancement: Complete conversation records are still preserved; when retrieval is triggered, more precise detail recall than humans is possible

Maintains natural human conversation focus characteristics while leveraging machine advantages in precise retrieval. During conversation, complete history isn't used; detailed retrieval is only activated under specific trigger conditions.

### Engineering Simulation Focus
```
Exclude unnecessary information: Remove from key summaries
Maintain focus: Use structured summaries, like mental overviews
Active recall: Automatically retrieve relevant historical content for each question
State updates: Continuous summarization, like mental event understanding
```
- No replay of complete content; instead use summaries to simulate human general overviews
- Summarize into new overviews to adjust conversation direction, simulating human mental perspectives
- Active retrieval of relevant history simulates human associative memory

#### Implementation
1. **Continuous mental perspective updates → Automatic summary updates** (Relevant info retention vs complete history)<br>
  After each conversation, humans unconsciously update their summary of current conversation based on new information and conduct next turn with new perspective
2. **Active associative memory → Fuzzy search system** (Automatic memory retrieval)<br>
  For each new question, automatically search relevant content in conversation history, simulating human active association of past discussions
3. **Current state focus → Fixed context framework** (Structured summaries)<br>
  Dynamically adjust current conversation focus instead of reviewing entire conversation history

| Cognitive Mode | Human Behavior | LLM | Simulation Implementation |
| - | - | - | - |
| **Memory Management** | Selective retention | Perfect recall | Structured forgetting |
| **Error Learning** | Avoid known failures | Repeat mistakes | Excluded options tracking |
| **Focus Maintenance** | Current state oriented | Historical drowning | Summary-based context |
| **Memory Retrieval** | Active associative triggering | Passive complete memory | Automatic fuzzy search |

## Fuzzy Retrieval Algorithm Design

### Multi-dimensional Scoring Mechanism
```
Total Score = Keyword Overlap(40%) + Semantic Similarity(40%) + Time Weight(20%)
```

**Keyword Overlap**
- Uses Jaccard similarity to calculate vocabulary matching degree
- Supports partial matching and containment relationships

**Semantic Similarity**
- Simplified cosine similarity calculating common vocabulary proportion
- Suitable for Chinese-English mixed text processing

**Time Weight**
- Linear decay within 24 hours: Recent=1.0, 24 hours ago=0.7
- Fixed score 0.7 after 24 hours (suitable for long-term continuous conversations)

### Retrieval Control Mechanism
- **Relevance Threshold**: Default 0.3, filters irrelevant content
- **Result Quantity Limit**: Maximum 5 most relevant records returned
- **Keyword Extraction**: Automatically filters stop words, retains meaningful vocabulary

### Context Combination Strategy
```
Each Turn Context = [Structured Summary] + [Relevant Historical Conversation] + [New Question]
```

## Implemented
- [x] **Structured Summary System**: Simulates human mental general overviews
- [x] **State Update Mechanism**: Automatically updates cognitive state after each conversation turn (gpt-4o-mini)
- [x] **Error Learning System**: Avoids repeated mistakes through `ExcludedOptions`
- [x] **Token Efficiency Optimization**: Fixed transmission of summaries and new content, no longer passing complete message strings
- [x] **Fuzzy Retrieval Mechanism**: Automatically retrieves relevant historical conversations as reference
- [x] **Multi-dimensional Scoring Algorithm**: Comprehensive relevance assessment of keywords+semantics+time
- [x] **Long Conversation Optimization**: Time weighting design suitable for continuous conversation scenarios

## To Be Implemented
- [ ] **Semantic Understanding Enhancement**: Integrate more precise semantic similarity algorithms
- [ ] **Keyword Extraction Optimization**: More intelligent vocabulary extraction and weight allocation
- [ ] **Dynamic Threshold Adjustment**: Automatically adjust relevance thresholds based on conversation content
- [ ] **Conversation Type Identification**: Optimize memory strategies for different conversation scenarios
- [ ] **Multi-model Support**: Support more LLM providers (Claude, Gemini, etc.)

## Example Usage

### Environment Requirements
- Go 1.20 or higher
- OpenAI API key

### Installation Steps

1. **Clone Project**
```bash
git clone https://github.com/pardnchiu/llm-dialogue-simulation 
cd llm-dialogue-simulation
```

2. **Configure API Key**
Create `OPENAI_API_KEY` file and put your OpenAI API key:
```bash
echo "your-openai-api-key-here" > OPENAI_API_KEY
```

Or set environment variable:
```bash
export OPENAI_API_KEY="your-openai-api-key-here"
```

3. **Run Program**
```bash
./llmsd
```
or
```bash
go run main.go
```

#### API Key Configuration
The program searches for OpenAI API key in this order:
1. Environment variable `OPENAI_API_KEY`
2. `OPENAI_API_KEY` file in current directory
3. `OPENAI_API_KEY` file in executable directory

#### Instruction File Configuration
**INSTRUCTION_CONVERSATION**
- Defines system instructions for main conversation model (GPT-4o)
- Affects AI assistant's response style and behavior
- Uses blank instructions if file doesn't exist

**INSTRUCTION_SUMMARY**
- Defines system instructions for summary generation model (GPT-4o-mini)
- Affects conversation summary update logic and format
- Uses blank instructions if file doesn't exist

### Usage

1. **Start Program**: After execution, displays three-panel interface
   - Left: Conversation history display
   - Top right: Conversation summary display
   - Bottom right: Question input field

2. **Basic Operations**:
   - `Enter`: Submit question
   - `Tab`: Switch panel focus
   - `Ctrl+C`: Exit program

3. **Conversation Flow**:
   - After inputting question, system automatically retrieves relevant conversation history
   - AI provides answers based on summary and relevant history
   - System automatically updates conversation summary, maintaining memory state (wait for summary update before continuing conversation)

## License

This source code project is licensed under the [MIT](LICENSE) license.

## Author

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)
