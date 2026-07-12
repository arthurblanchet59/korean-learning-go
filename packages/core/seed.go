package core

import "time"

type seedCardSpec struct {
	id                 string
	deckID             string
	kind               CardKind
	korean             string
	translation        string
	romanization       string
	exampleKorean      string
	exampleTranslation string
	tags               []string
}

func SeedDeck(now time.Time) Deck {
	return SeedDecks(now)[0]
}

func SeedDecks(now time.Time) []Deck {
	return []Deck{
		{ID: "starter", Name: "Essentiels coreens", Description: "Salutations, personnes et mots indispensables pour commencer.", CreatedAt: now},
		{ID: "daily", Name: "Vie quotidienne", Description: "Temps, lieux et objets rencontres chaque jour.", CreatedAt: now},
		{ID: "food", Name: "Nourriture et boissons", Description: "Commander, cuisiner et parler de ses gouts.", CreatedAt: now},
		{ID: "verbs", Name: "Verbes fondamentaux", Description: "Les actions les plus utiles, avec leur forme polie en contexte.", CreatedAt: now},
	}
}

func SeedCards(now time.Time) []Card {
	specs := []seedCardSpec{
		{id: "starter-1", deckID: "starter", kind: CardKindPhrase, korean: "안녕하세요", translation: "bonjour", romanization: "annyeonghaseyo", exampleKorean: "안녕하세요, 저는 아서예요.", exampleTranslation: "Bonjour, je suis Arthur.", tags: []string{"salutation", "politesse"}},
		{id: "starter-2", deckID: "food", kind: CardKindVocabulary, korean: "물", translation: "eau", romanization: "mul", exampleKorean: "물을 마셔요.", exampleTranslation: "Je bois de l'eau.", tags: []string{"boisson", "base"}},
		{id: "starter-3", deckID: "daily", kind: CardKindVocabulary, korean: "학교", translation: "école", romanization: "hakgyo", exampleKorean: "학교에 가요.", exampleTranslation: "Je vais à l'école.", tags: []string{"lieu", "base"}},
		{id: "starter-yes", deckID: "starter", kind: CardKindPhrase, korean: "네", translation: "oui", romanization: "ne", exampleKorean: "네, 좋아요.", exampleTranslation: "Oui, c'est bien.", tags: []string{"réponse", "base"}},
		{id: "starter-no", deckID: "starter", kind: CardKindPhrase, korean: "아니요", translation: "non", romanization: "aniyo", exampleKorean: "아니요, 괜찮아요.", exampleTranslation: "Non, ça va.", tags: []string{"réponse", "base"}},
		{id: "starter-thanks", deckID: "starter", kind: CardKindPhrase, korean: "감사합니다", translation: "merci", romanization: "gamsahamnida", exampleKorean: "도와주셔서 감사합니다.", exampleTranslation: "Merci de m'avoir aidé.", tags: []string{"politesse", "salutation"}},
		{id: "starter-sorry", deckID: "starter", kind: CardKindPhrase, korean: "죄송합니다", translation: "désolé", romanization: "joesonghamnida", exampleKorean: "늦어서 죄송합니다.", exampleTranslation: "Désolé d'être en retard.", tags: []string{"politesse", "excuse"}},
		{id: "starter-please", deckID: "starter", kind: CardKindPhrase, korean: "주세요", translation: "donnez-moi s'il vous plaît", romanization: "juseyo", exampleKorean: "커피 한 잔 주세요.", exampleTranslation: "Un café, s'il vous plaît.", tags: []string{"politesse", "commande"}},
		{id: "starter-person", deckID: "starter", kind: CardKindVocabulary, korean: "사람", translation: "personne", romanization: "saram", exampleKorean: "그 사람은 제 친구예요.", exampleTranslation: "Cette personne est mon ami.", tags: []string{"personne", "base"}},
		{id: "starter-friend", deckID: "starter", kind: CardKindVocabulary, korean: "친구", translation: "ami", romanization: "chingu", exampleKorean: "친구를 만나요.", exampleTranslation: "Je rencontre un ami.", tags: []string{"personne", "relation"}},
		{id: "starter-family", deckID: "starter", kind: CardKindVocabulary, korean: "가족", translation: "famille", romanization: "gajok", exampleKorean: "가족과 같이 살아요.", exampleTranslation: "Je vis avec ma famille.", tags: []string{"personne", "famille"}},
		{id: "starter-home", deckID: "starter", kind: CardKindVocabulary, korean: "집", translation: "maison", romanization: "jip", exampleKorean: "집에 있어요.", exampleTranslation: "Je suis à la maison.", tags: []string{"lieu", "base"}},
		{id: "starter-name", deckID: "starter", kind: CardKindVocabulary, korean: "이름", translation: "nom", romanization: "ireum", exampleKorean: "이름이 뭐예요?", exampleTranslation: "Comment vous appelez-vous ?", tags: []string{"identité", "base"}},
		{id: "starter-korea", deckID: "starter", kind: CardKindVocabulary, korean: "한국", translation: "Corée", romanization: "hanguk", exampleKorean: "한국에 가고 싶어요.", exampleTranslation: "Je veux aller en Corée.", tags: []string{"pays", "culture"}},

		{id: "daily-today", deckID: "daily", kind: CardKindVocabulary, korean: "오늘", translation: "aujourd'hui", romanization: "oneul", exampleKorean: "오늘은 바빠요.", exampleTranslation: "Aujourd'hui, je suis occupé.", tags: []string{"temps", "jour"}},
		{id: "daily-tomorrow", deckID: "daily", kind: CardKindVocabulary, korean: "내일", translation: "demain", romanization: "naeil", exampleKorean: "내일 만나요.", exampleTranslation: "On se voit demain.", tags: []string{"temps", "jour"}},
		{id: "daily-yesterday", deckID: "daily", kind: CardKindVocabulary, korean: "어제", translation: "hier", romanization: "eoje", exampleKorean: "어제 영화를 봤어요.", exampleTranslation: "Hier, j'ai regardé un film.", tags: []string{"temps", "jour"}},
		{id: "daily-now", deckID: "daily", kind: CardKindVocabulary, korean: "지금", translation: "maintenant", romanization: "jigeum", exampleKorean: "지금 공부해요.", exampleTranslation: "J'étudie maintenant.", tags: []string{"temps", "base"}},
		{id: "daily-time", deckID: "daily", kind: CardKindVocabulary, korean: "시간", translation: "temps", romanization: "sigan", exampleKorean: "시간이 없어요.", exampleTranslation: "Je n'ai pas le temps.", tags: []string{"temps", "nom"}},
		{id: "daily-morning", deckID: "daily", kind: CardKindVocabulary, korean: "아침", translation: "matin", romanization: "achim", exampleKorean: "아침에 커피를 마셔요.", exampleTranslation: "Je bois du café le matin.", tags: []string{"temps", "journée"}},
		{id: "daily-evening", deckID: "daily", kind: CardKindVocabulary, korean: "저녁", translation: "soir", romanization: "jeonyeok", exampleKorean: "저녁을 먹어요.", exampleTranslation: "Je mange le dîner.", tags: []string{"temps", "journée"}},
		{id: "daily-company", deckID: "daily", kind: CardKindVocabulary, korean: "회사", translation: "entreprise", romanization: "hoesa", exampleKorean: "회사에서 일해요.", exampleTranslation: "Je travaille dans une entreprise.", tags: []string{"lieu", "travail"}},
		{id: "daily-book", deckID: "daily", kind: CardKindVocabulary, korean: "책", translation: "livre", romanization: "chaek", exampleKorean: "한국어 책을 읽어요.", exampleTranslation: "Je lis un livre de coréen.", tags: []string{"objet", "étude"}},
		{id: "daily-phone", deckID: "daily", kind: CardKindVocabulary, korean: "휴대폰", translation: "téléphone portable", romanization: "hyudaepon", exampleKorean: "휴대폰을 찾아요.", exampleTranslation: "Je cherche mon téléphone.", tags: []string{"objet", "technologie"}},
		{id: "daily-subway", deckID: "daily", kind: CardKindVocabulary, korean: "지하철", translation: "métro", romanization: "jihacheol", exampleKorean: "지하철을 타요.", exampleTranslation: "Je prends le métro.", tags: []string{"transport", "ville"}},
		{id: "daily-station", deckID: "daily", kind: CardKindVocabulary, korean: "역", translation: "gare", romanization: "yeok", exampleKorean: "서울역은 어디예요?", exampleTranslation: "Où est la gare de Séoul ?", tags: []string{"transport", "lieu"}},
		{id: "daily-bathroom", deckID: "daily", kind: CardKindVocabulary, korean: "화장실", translation: "toilettes", romanization: "hwajangsil", exampleKorean: "화장실이 어디예요?", exampleTranslation: "Où sont les toilettes ?", tags: []string{"lieu", "utile"}},
		{id: "daily-restaurant", deckID: "daily", kind: CardKindVocabulary, korean: "식당", translation: "restaurant", romanization: "sikdang", exampleKorean: "이 식당은 맛있어요.", exampleTranslation: "Ce restaurant est bon.", tags: []string{"lieu", "repas"}},
		{id: "daily-cafe", deckID: "daily", kind: CardKindVocabulary, korean: "카페", translation: "café", romanization: "kape", exampleKorean: "카페에서 친구를 만나요.", exampleTranslation: "Je retrouve un ami au café.", tags: []string{"lieu", "sortie"}},
		{id: "daily-store", deckID: "daily", kind: CardKindVocabulary, korean: "가게", translation: "magasin", romanization: "gage", exampleKorean: "가게에서 빵을 사요.", exampleTranslation: "J'achète du pain au magasin.", tags: []string{"lieu", "achat"}},
		{id: "daily-money", deckID: "daily", kind: CardKindVocabulary, korean: "돈", translation: "argent", romanization: "don", exampleKorean: "돈이 조금 있어요.", exampleTranslation: "J'ai un peu d'argent.", tags: []string{"achat", "nom"}},

		{id: "food-rice", deckID: "food", kind: CardKindVocabulary, korean: "밥", translation: "riz cuit ou repas", romanization: "bap", exampleKorean: "밥을 먹었어요?", exampleTranslation: "Avez-vous mangé ?", tags: []string{"repas", "base"}},
		{id: "food-food", deckID: "food", kind: CardKindVocabulary, korean: "음식", translation: "nourriture", romanization: "eumsik", exampleKorean: "한국 음식을 좋아해요.", exampleTranslation: "J'aime la cuisine coréenne.", tags: []string{"repas", "base"}},
		{id: "food-coffee", deckID: "food", kind: CardKindVocabulary, korean: "커피", translation: "café", romanization: "keopi", exampleKorean: "커피 한 잔 주세요.", exampleTranslation: "Un café, s'il vous plaît.", tags: []string{"boisson", "commande"}},
		{id: "food-tea", deckID: "food", kind: CardKindVocabulary, korean: "차", translation: "thé", romanization: "cha", exampleKorean: "따뜻한 차를 마셔요.", exampleTranslation: "Je bois du thé chaud.", tags: []string{"boisson", "base"}},
		{id: "food-milk", deckID: "food", kind: CardKindVocabulary, korean: "우유", translation: "lait", romanization: "uyu", exampleKorean: "우유가 냉장고에 있어요.", exampleTranslation: "Le lait est dans le réfrigérateur.", tags: []string{"boisson", "base"}},
		{id: "food-bread", deckID: "food", kind: CardKindVocabulary, korean: "빵", translation: "pain", romanization: "ppang", exampleKorean: "아침에 빵을 먹어요.", exampleTranslation: "Je mange du pain le matin.", tags: []string{"repas", "base"}},
		{id: "food-apple", deckID: "food", kind: CardKindVocabulary, korean: "사과", translation: "pomme", romanization: "sagwa", exampleKorean: "사과 두 개를 샀어요.", exampleTranslation: "J'ai acheté deux pommes.", tags: []string{"fruit", "aliment"}},
		{id: "food-fruit", deckID: "food", kind: CardKindVocabulary, korean: "과일", translation: "fruit", romanization: "gwail", exampleKorean: "과일을 자주 먹어요.", exampleTranslation: "Je mange souvent des fruits.", tags: []string{"fruit", "aliment"}},
		{id: "food-meat", deckID: "food", kind: CardKindVocabulary, korean: "고기", translation: "viande", romanization: "gogi", exampleKorean: "고기를 안 먹어요.", exampleTranslation: "Je ne mange pas de viande.", tags: []string{"aliment", "repas"}},
		{id: "food-fish", deckID: "food", kind: CardKindVocabulary, korean: "생선", translation: "poisson", romanization: "saengseon", exampleKorean: "생선을 구워요.", exampleTranslation: "Je fais griller du poisson.", tags: []string{"aliment", "repas"}},
		{id: "food-vegetable", deckID: "food", kind: CardKindVocabulary, korean: "채소", translation: "légume", romanization: "chaeso", exampleKorean: "채소를 많이 먹어요.", exampleTranslation: "Je mange beaucoup de légumes.", tags: []string{"aliment", "santé"}},
		{id: "food-egg", deckID: "food", kind: CardKindVocabulary, korean: "계란", translation: "œuf", romanization: "gyeran", exampleKorean: "계란을 두 개 넣어요.", exampleTranslation: "Je mets deux œufs.", tags: []string{"aliment", "cuisine"}},
		{id: "food-delicious", deckID: "food", kind: CardKindVocabulary, korean: "맛있어요", translation: "c'est délicieux", romanization: "masisseoyo", exampleKorean: "이 비빔밥은 정말 맛있어요.", exampleTranslation: "Ce bibimbap est vraiment délicieux.", tags: []string{"goût", "adjectif"}},
		{id: "food-hungry", deckID: "food", kind: CardKindPhrase, korean: "배고파요", translation: "j'ai faim", romanization: "baegopayo", exampleKorean: "지금 배고파요.", exampleTranslation: "J'ai faim maintenant.", tags: []string{"sensation", "phrase"}},
		{id: "food-thirsty", deckID: "food", kind: CardKindPhrase, korean: "목말라요", translation: "j'ai soif", romanization: "mongmallayo", exampleKorean: "운동 후에 목말라요.", exampleTranslation: "J'ai soif après le sport.", tags: []string{"sensation", "phrase"}},

		{id: "verb-go", deckID: "verbs", kind: CardKindVocabulary, korean: "가다", translation: "aller", romanization: "gada", exampleKorean: "매일 학교에 가요.", exampleTranslation: "Je vais à l'école tous les jours.", tags: []string{"verbe", "déplacement"}},
		{id: "verb-come", deckID: "verbs", kind: CardKindVocabulary, korean: "오다", translation: "venir", romanization: "oda", exampleKorean: "친구가 집에 와요.", exampleTranslation: "Un ami vient à la maison.", tags: []string{"verbe", "déplacement"}},
		{id: "verb-eat", deckID: "verbs", kind: CardKindVocabulary, korean: "먹다", translation: "manger", romanization: "meokda", exampleKorean: "점심을 같이 먹어요.", exampleTranslation: "Nous déjeunons ensemble.", tags: []string{"verbe", "repas"}},
		{id: "verb-drink", deckID: "verbs", kind: CardKindVocabulary, korean: "마시다", translation: "boire", romanization: "masida", exampleKorean: "물을 많이 마셔요.", exampleTranslation: "Je bois beaucoup d'eau.", tags: []string{"verbe", "repas"}},
		{id: "verb-see", deckID: "verbs", kind: CardKindVocabulary, korean: "보다", translation: "voir ou regarder", romanization: "boda", exampleKorean: "주말에 영화를 봐요.", exampleTranslation: "Je regarde un film le week-end.", tags: []string{"verbe", "perception"}},
		{id: "verb-read", deckID: "verbs", kind: CardKindVocabulary, korean: "읽다", translation: "lire", romanization: "ikda", exampleKorean: "책을 천천히 읽어요.", exampleTranslation: "Je lis le livre lentement.", tags: []string{"verbe", "étude"}},
		{id: "verb-write", deckID: "verbs", kind: CardKindVocabulary, korean: "쓰다", translation: "écrire", romanization: "sseuda", exampleKorean: "한국어로 일기를 써요.", exampleTranslation: "J'écris un journal en coréen.", tags: []string{"verbe", "étude"}},
		{id: "verb-speak", deckID: "verbs", kind: CardKindVocabulary, korean: "말하다", translation: "parler", romanization: "malhada", exampleKorean: "한국어로 말해요.", exampleTranslation: "Je parle en coréen.", tags: []string{"verbe", "communication"}},
		{id: "verb-listen", deckID: "verbs", kind: CardKindVocabulary, korean: "듣다", translation: "écouter", romanization: "deutda", exampleKorean: "한국 음악을 들어요.", exampleTranslation: "J'écoute de la musique coréenne.", tags: []string{"verbe", "perception"}},
		{id: "verb-study", deckID: "verbs", kind: CardKindVocabulary, korean: "공부하다", translation: "étudier", romanization: "gongbuhada", exampleKorean: "매일 한국어를 공부해요.", exampleTranslation: "J'étudie le coréen chaque jour.", tags: []string{"verbe", "étude"}},
		{id: "verb-work", deckID: "verbs", kind: CardKindVocabulary, korean: "일하다", translation: "travailler", romanization: "ilhada", exampleKorean: "회사에서 일해요.", exampleTranslation: "Je travaille dans une entreprise.", tags: []string{"verbe", "travail"}},
		{id: "verb-sleep", deckID: "verbs", kind: CardKindVocabulary, korean: "자다", translation: "dormir", romanization: "jada", exampleKorean: "열한 시에 자요.", exampleTranslation: "Je dors à onze heures.", tags: []string{"verbe", "routine"}},
		{id: "verb-buy", deckID: "verbs", kind: CardKindVocabulary, korean: "사다", translation: "acheter", romanization: "sada", exampleKorean: "시장에서 과일을 사요.", exampleTranslation: "J'achète des fruits au marché.", tags: []string{"verbe", "achat"}},
		{id: "verb-meet", deckID: "verbs", kind: CardKindVocabulary, korean: "만나다", translation: "rencontrer", romanization: "mannada", exampleKorean: "역에서 친구를 만나요.", exampleTranslation: "Je retrouve un ami à la gare.", tags: []string{"verbe", "relation"}},
		{id: "verb-like", deckID: "verbs", kind: CardKindVocabulary, korean: "좋아하다", translation: "aimer", romanization: "joahada", exampleKorean: "저는 커피를 좋아해요.", exampleTranslation: "J'aime le café.", tags: []string{"verbe", "goût"}},
		{id: "verb-have", deckID: "verbs", kind: CardKindVocabulary, korean: "있다", translation: "avoir ou être présent", romanization: "itda", exampleKorean: "질문이 있어요.", exampleTranslation: "J'ai une question.", tags: []string{"verbe", "existence"}},
		{id: "verb-do", deckID: "verbs", kind: CardKindVocabulary, korean: "하다", translation: "faire", romanization: "hada", exampleKorean: "주말에 운동을 해요.", exampleTranslation: "Je fais du sport le week-end.", tags: []string{"verbe", "base"}},
		{id: "verb-know", deckID: "verbs", kind: CardKindVocabulary, korean: "알다", translation: "savoir ou connaître", romanization: "alda", exampleKorean: "그 사람을 알아요.", exampleTranslation: "Je connais cette personne.", tags: []string{"verbe", "connaissance"}},
		{id: "verb-not-know", deckID: "verbs", kind: CardKindPhrase, korean: "모르겠어요", translation: "je ne sais pas", romanization: "moreugesseoyo", exampleKorean: "미안하지만 모르겠어요.", exampleTranslation: "Désolé, mais je ne sais pas.", tags: []string{"phrase", "utile"}},
		{id: "verb-wait", deckID: "verbs", kind: CardKindVocabulary, korean: "기다리다", translation: "attendre", romanization: "gidarida", exampleKorean: "여기에서 기다려 주세요.", exampleTranslation: "Veuillez attendre ici.", tags: []string{"verbe", "utile"}},
	}

	cards := make([]Card, 0, len(specs))
	for index, spec := range specs {
		state := NewState(now)
		state.NextReviewAt = now.AddDate(0, 0, index/12)
		cards = append(cards, Card{
			ID:                 spec.id,
			DeckID:             spec.deckID,
			Kind:               spec.kind,
			Korean:             spec.korean,
			Translation:        spec.translation,
			Romanization:       spec.romanization,
			ExampleKorean:      spec.exampleKorean,
			ExampleTranslation: spec.exampleTranslation,
			Tags:               spec.tags,
			CreatedAt:          now,
			ReviewState:        state,
		})
	}
	return cards
}
