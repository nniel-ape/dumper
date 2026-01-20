package llm

const ProcessContentPrompt = `Analyze the following content and extract structured information.
%s
Content Type: %s
Content:
---
%s
---

Respond with ONLY valid JSON (no markdown, no explanation):
{
  "title": "concise descriptive title (max 10 words)",
  "summary": "2-3 sentence summary capturing key points",
  "tags": ["tag1", "tag2", "tag3"],
  "related_topics": ["topic that might connect to other saved items"]
}

Rules:
- Tags should be lowercase, single words or short phrases
- If the content includes hashtags (e.g. #tag), include them as tags without the "#"
- Generate 3-7 relevant tags
- PREFER reusing existing tags when they fit the content (consistency is valuable)
- Summary should be informative but concise
- Related topics help build knowledge graph connections`

const FindRelationshipsPrompt = `Given a new item and existing items, identify ONLY genuinely related items.

New item:
Title: %s
Summary: %s
Tags: %v

Existing items:
%s

Respond with ONLY valid JSON array of relationships:
[
  {"target_id": "id", "relation_type": "type", "strength": 0.8}
]

Relation types: "similar_topic", "references", "contradicts", "extends", "prerequisite"
Strength: 0.7-1.0 (only strong, obvious connections)

STRICT RULES - read carefully:
1. Items MUST be in the same knowledge domain (e.g., both about programming, both about cinema, both about cooking)
2. Do NOT connect items just because they share generic tags like "technology", "article", "image"
3. Do NOT use "creative interpretation" - the connection must be obvious to any reader
4. When in doubt, DO NOT create a relationship - return an empty array []
5. Quality over quantity: 0-2 relationships is normal, more than 3 is suspicious

INVALID relationships (never create these):
- Film about funeral rites → Programming framework (different domains)
- Random image → Software tool (no semantic connection)
- Russian literature → JavaScript library (unrelated)
- Cooking recipe → Database article (unrelated)

VALID relationships:
- Go tutorial → Go framework documentation (same language/domain)
- React hooks article → React state management guide (same framework)
- Film review → Director's biography (same domain: cinema)

Return empty array [] if no strong relationships exist.`

const AnswerQuestionPrompt = `Based on the following saved knowledge items, answer the user's question.

User question: %s

Relevant items:
%s

Instructions:
- Synthesize information from the provided items
- Be concise but informative
- If the answer is not in the provided items, say so
- Reference specific items when relevant

Answer:`

const SummarizeSearchPrompt = `You are a knowledge assistant. The user searched for a topic and we found some information.
Create a helpful knowledge entry about this topic.
%s
Topic: %s

Search Results:
---
%s
---

Respond with ONLY valid JSON (no markdown, no explanation):
{
  "title": "concise descriptive title for this topic (max 10 words)",
  "summary": "2-4 sentence informative summary about this topic",
  "tags": ["tag1", "tag2", "tag3"],
  "related_topics": ["related topic 1", "related topic 2"]
}

Rules:
- If search results are empty or unhelpful, use your general knowledge about the topic
- Tags should be lowercase, relevant to the topic
- Generate 3-5 relevant tags
- PREFER reusing existing tags when they fit the topic (consistency is valuable)
- Summary should explain what this topic is and why it's notable
- Include the most important facts or uses`
