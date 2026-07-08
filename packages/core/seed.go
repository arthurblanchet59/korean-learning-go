package core

import "time"

func SeedDeck(now time.Time) Deck {
	return Deck{
		ID:          "deck-basic-korean",
		Name:        "Coreen debutant",
		Description: "Premiers mots et phrases utiles.",
		CreatedAt:   now,
	}
}

func SeedCards(now time.Time) []Card {
	state := NewState(now)

	return []Card{
		{
			ID:                 "card-hello",
			DeckID:             "deck-basic-korean",
			Kind:               CardKindPhrase,
			Korean:             "안녕하세요",
			Translation:        "bonjour",
			Romanization:       "annyeonghaseyo",
			ExampleKorean:      "안녕하세요, 저는 아서예요.",
			ExampleTranslation: "Bonjour, je suis Arthur.",
			Tags:               []string{"salutation", "politesse"},
			CreatedAt:          now,
			ReviewState:        state,
		},
		{
			ID:                 "card-water",
			DeckID:             "deck-basic-korean",
			Kind:               CardKindVocabulary,
			Korean:             "물",
			Translation:        "eau",
			Romanization:       "mul",
			ExampleKorean:      "물을 마셔요.",
			ExampleTranslation: "Je bois de l'eau.",
			Tags:               []string{"base", "nourriture"},
			CreatedAt:          now,
			ReviewState:        state,
		},
		{
			ID:                 "card-school",
			DeckID:             "deck-basic-korean",
			Kind:               CardKindVocabulary,
			Korean:             "학교",
			Translation:        "ecole",
			Romanization:       "hakgyo",
			ExampleKorean:      "학교에 가요.",
			ExampleTranslation: "Je vais a l'ecole.",
			Tags:               []string{"lieu", "base"},
			CreatedAt:          now,
			ReviewState:        state,
		},
	}
}
