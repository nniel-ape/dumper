package llm

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type ProcessedContent struct {
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Tags          []string `json:"tags"`
	RelatedTopics []string `json:"related_topics"`
}

type RelationshipSuggestion struct {
	TargetID     string  `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Strength     float64 `json:"strength"`
}
