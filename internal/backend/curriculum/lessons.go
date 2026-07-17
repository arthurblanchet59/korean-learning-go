package curriculum

import (
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func Lessons() []core.Lesson {
	return []core.Lesson{
		{ID: "hangul-1", Title: "Les voyelles du hangeul", Description: "Lire et distinguer les dix voyelles simples.", Level: "A0", Order: 1, Content: `OBJECTIF
Reconnaître les voyelles sans dépendre de la romanisation et comprendre leur orientation dans un bloc.

RÈGLE
Les voyelles verticales se placent à droite de la consonne : ㅏ a, ㅓ eo, ㅣ i. Les voyelles horizontales se placent dessous : ㅗ o, ㅜ u, ㅡ eu.

Les quatre voyelles avec un petit trait supplémentaire commencent par le son y : ㅑ ya, ㅕ yeo, ㅛ yo, ㅠ yu.

EXEMPLES
아 a · 어 eo · 오 o · 우 u · 으 eu · 이 i
야 ya · 여 yeo · 요 yo · 유 yu

À RETENIR
ㅓ n'est pas le son français « o ». C'est un son ouvert, souvent noté eo. ㅡ se prononce avec les lèvres non arrondies.

PRATIQUE
Lis à voix haute : 아, 어, 오, 우, 으, 이, 야, 여, 요, 유. Puis écris la romanisation de 여, 우 et 으.

CORRIGÉ
여 = yeo, 우 = u, 으 = eu.`},
		{ID: "hangul-consonants", Title: "Les consonnes de base", Description: "Reconnaître les consonnes et leur son selon la position.", Level: "A0", Order: 2, Content: `OBJECTIF
Lire les quatorze consonnes simples et éviter de leur attribuer un son français trop rigide.

RÈGLE
ㄱ g/k · ㄴ n · ㄷ d/t · ㄹ r/l · ㅁ m · ㅂ b/p · ㅅ s · ㅇ muet/ng · ㅈ j · ㅊ ch · ㅋ k · ㅌ t · ㅍ p · ㅎ h.

Au début d'une syllabe, ㅇ est muet : 아 se lit a. En fin de syllabe, il se prononce ng : 강 se lit gang.

EXEMPLES
가 ga · 나 na · 다 da · 라 ra · 마 ma · 바 ba · 사 sa · 자 ja · 차 cha · 카 ka · 타 ta · 파 pa · 하 ha

À RETENIR
ㄹ ressemble à un r léger entre deux voyelles, mais se rapproche de l en fin de syllabe. ㄱ, ㄷ et ㅂ sont moins fortement voisées que g, d et b en français.

PRATIQUE
Lis : 나라, 바다, 사람, 하나. Entoure mentalement la consonne initiale de chaque bloc.

CORRIGÉ
나라 nara, 바다 bada, 사람 saram, 하나 hana.`},
		{ID: "hangul-2", Title: "Former un bloc syllabique", Description: "Assembler consonne, voyelle et consonne finale.", Level: "A0", Order: 3, Content: `OBJECTIF
Comprendre pourquoi les lettres coréennes sont regroupées en carrés et savoir les décomposer.

RÈGLE
Un bloc contient au minimum une consonne initiale et une voyelle. Avec une voyelle verticale : ㄱ + ㅏ = 가. Avec une voyelle horizontale : ㄱ + ㅗ = 고.

Une consonne finale, appelée 받침 batchim, se place sous le bloc : ㄱ + ㅏ + ㄴ = 간.

EXEMPLES
나 = ㄴ + ㅏ · 너 = ㄴ + ㅓ · 노 = ㄴ + ㅗ · 누 = ㄴ + ㅜ
한 = ㅎ + ㅏ + ㄴ · 국 = ㄱ + ㅜ + ㄱ · 밥 = ㅂ + ㅏ + ㅂ

À RETENIR
Lis toujours dans cet ordre : initiale, voyelle, finale. La forme carrée ne change pas cet ordre.

PRATIQUE
Décompose 한글, 사람 et 물 en lettres individuelles.

CORRIGÉ
한 = ㅎ+ㅏ+ㄴ, 글 = ㄱ+ㅡ+ㄹ. 사 = ㅅ+ㅏ, 람 = ㄹ+ㅏ+ㅁ. 물 = ㅁ+ㅜ+ㄹ.`},
		{ID: "hangul-batchim", Title: "Comprendre le batchim", Description: "Prononcer les consonnes finales et choisir les particules.", Level: "A0", Order: 4, Content: `OBJECTIF
Repérer une consonne finale et comprendre son influence sur la prononciation et la grammaire.

RÈGLE
En fin de bloc, de nombreuses consonnes se réduisent à sept sons : ㄱ k, ㄴ n, ㄷ t, ㄹ l, ㅁ m, ㅂ p, ㅇ ng.

Le batchim décide aussi de nombreuses particules. Après une consonne finale : 은, 이, 을. Après une voyelle : 는, 가, 를.

EXEMPLES
밥 se termine par le son p · 책 par k · 옷 par t · 물 par l · 방 par ng.
책은 mais 커피는 · 물이 mais 학교가 · 밥을 mais 사과를.

À RETENIR
Observe la dernière case du bloc, pas la dernière lettre de la romanisation.

PRATIQUE
Choisis la bonne forme : 집(은/는), 친구(이/가), 책(을/를), 학교(은/는).

CORRIGÉ
집은, 친구가, 책을, 학교는.`},
		{ID: "hangul-sounds", Title: "Les liaisons utiles", Description: "Lire plus naturellement quand les syllabes se rencontrent.", Level: "A0", Order: 5, Content: `OBJECTIF
Reconnaître les changements de son les plus fréquents sans apprendre toutes les règles phonétiques d'un coup.

RÈGLE
Quand un batchim est suivi d'une syllabe commençant par ㅇ, son son passe souvent à la syllabe suivante. 한국어 se rapproche de 한구거 et 먹어요 de 머거요.

ㄴ et ㄹ s'influencent souvent : 신라 se prononce 실라. Une consonne peut aussi se renforcer : 학교 se rapproche de 학꾜.

EXEMPLES
한국어 hangugeo · 먹어요 meogeoyo · 음악 eumak · 집에 jibe.

À RETENIR
L'orthographe ne change pas. Ces règles expliquent seulement la prononciation réelle.

PRATIQUE
Lis lentement puis naturellement : 한국어를 배워요. 음악을 들어요. 집에 있어요.

CORRIGÉ
Prononciation approchée : 한구거를 배워요, 으마글 드러요, 지베 이써요.`},
		{ID: "grammar-present", Title: "Le présent poli en -아요/-어요", Description: "Conjuguer les verbes courants dans une conversation.", Level: "A1", Order: 6, Content: `OBJECTIF
Transformer la forme du dictionnaire en une forme polie utilisable au quotidien.

RÈGLE
Retire 다. Si la dernière voyelle du radical est ㅏ ou ㅗ, ajoute 아요. Sinon, ajoute 어요. 하다 devient 해요.

Des contractions sont naturelles : 가다 → 가요, 보다 → 봐요, 마시다 → 마셔요, 배우다 → 배워요.

EXEMPLES
먹다 → 먹어요 · 읽다 → 읽어요 · 오다 → 와요 · 공부하다 → 공부해요.

À RETENIR
Le coréen omet souvent le sujet lorsqu'il est évident. 커피를 마셔요 peut signifier « je bois du café » selon le contexte.

PRATIQUE
Conjugue : 자다, 쓰다, 만나다, 일하다. Puis traduis « J'étudie le coréen ».

CORRIGÉ
자요, 써요, 만나요, 일해요. 한국어를 공부해요.`},
		{ID: "grammar-topic", Title: "Thème 은/는 et sujet 이/가", Description: "Comprendre la différence plutôt que mémoriser deux paires.", Level: "A1", Order: 7, Content: `OBJECTIF
Présenter un thème avec 은/는 et identifier une information nouvelle avec 이/가.

RÈGLE
은/는 annonce ce dont on va parler ou crée un contraste. 이/가 marque souvent le sujet précis, nouveau ou mis en valeur.

Après batchim : 은 et 이. Après voyelle : 는 et 가.

EXEMPLES
저는 학생이에요. Quant à moi, je suis étudiant.
누가 와요? 민수가 와요. Qui vient ? Minsu vient.
커피는 좋아하지만 차는 안 좋아해요. J'aime le café, mais pas le thé.

À RETENIR
Il n'existe pas toujours une traduction française visible. Demande-toi si tu poses le décor (은/는) ou identifies le sujet (이/가).

PRATIQUE
Complète : 저__ 프랑스 사람이에요. 뭐__ 맛있어요? 이 책__ 재미있어요.

CORRIGÉ
저는 프랑스 사람이에요. 뭐가 맛있어요? 이 책은 재미있어요.`},
		{ID: "grammar-object", Title: "L'objet avec 을/를", Description: "Construire clairement une phrase sujet-objet-verbe.", Level: "A1", Order: 8, Content: `OBJECTIF
Marquer ce qui reçoit directement l'action du verbe.

RÈGLE
Ajoute 을 après un batchim et 를 après une voyelle. Le verbe se place généralement à la fin de la phrase.

Structure utile : thème + complément + objet + verbe. 저는 카페에서 커피를 마셔요.

EXEMPLES
책을 읽어요. Je lis un livre.
사과를 먹어요. Je mange une pomme.
한국어를 공부해요. J'étudie le coréen.

À RETENIR
À l'oral, 을/를 peut disparaître si le sens reste évident, mais utilise-le pendant l'apprentissage pour voir la structure.

PRATIQUE
Ajoute la particule puis conjugue : 물 + 마시다, 영화 + 보다, 친구 + 만나다.

CORRIGÉ
물을 마셔요. 영화를 봐요. 친구를 만나요.`},
		{ID: "grammar-place", Title: "Les lieux avec 에 et 에서", Description: "Distinguer destination, position et lieu d'une action.", Level: "A1", Order: 9, Content: `OBJECTIF
Choisir entre 에 et 에서 sans traduire mécaniquement le mot « à ».

RÈGLE
에 indique une destination, une position ou un moment : 학교에 가요, 집에 있어요, 세 시에 만나요.

에서 indique le lieu où une action se déroule : 학교에서 공부해요, 식당에서 먹어요.

EXEMPLES
서울에 가요. Je vais à Séoul.
서울에서 살아요. J'habite à Séoul.
카페에서 친구를 만나요. Je rencontre un ami au café.

À RETENIR
Avec 있다/없다 et les verbes de déplacement, pense d'abord à 에. Pour une activité réalisée sur place, pense à 에서.

PRATIQUE
Choisis : 회사(에/에서) 일해요. 집(에/에서) 와요. 일곱 시(에/에서) 자요.

CORRIGÉ
회사에서 일해요. 집에 와요. 일곱 시에 자요.`},
		{ID: "grammar-negation", Title: "Dire ne pas avec 안 et 못", Description: "Exprimer un choix négatif ou une impossibilité.", Level: "A1", Order: 10, Content: `OBJECTIF
Former des phrases négatives simples et distinguer volonté et capacité.

RÈGLE
안 devant le verbe signifie « ne pas ». 못 signifie « ne pas pouvoir ». Avec 하다, place-les avant 하다 : 공부 안 해요, 운동 못 해요.

La forme longue -지 않아요 est plus neutre ou écrite : 먹지 않아요.

EXEMPLES
고기를 안 먹어요. Je ne mange pas de viande.
오늘 못 가요. Je ne peux pas venir aujourd'hui.
커피를 좋아하지 않아요. Je n'aime pas le café.

À RETENIR
안 가요 = je n'y vais pas. 못 가요 = je ne peux pas y aller.

PRATIQUE
Traduis : « Je ne travaille pas aujourd'hui » et « Je ne peux pas lire ce livre ».

CORRIGÉ
오늘 일 안 해요. 이 책을 못 읽어요.`},
		{ID: "grammar-past", Title: "Le passé en -았어요/-었어요", Description: "Raconter une action terminée simplement.", Level: "A1", Order: 11, Content: `OBJECTIF
Parler d'hier, d'une expérience ou d'une action terminée.

RÈGLE
Le principe suit le présent : radical avec ㅏ/ㅗ + 았어요, autres voyelles + 었어요, 하다 → 했어요.

Les contractions sont fréquentes : 가다 → 갔어요, 보다 → 봤어요, 마시다 → 마셨어요.

EXEMPLES
어제 공부했어요. J'ai étudié hier.
친구를 만났어요. J'ai rencontré un ami.
비빔밥을 먹었어요. J'ai mangé un bibimbap.

À RETENIR
Le mot de temps peut être placé au début : 어제, 지난주에, 아침에.

PRATIQUE
Mets au passé : 가요, 읽어요, 자요, 일해요.

CORRIGÉ
갔어요, 읽었어요, 잤어요, 일했어요.`},
		{ID: "numbers-counters", Title: "Nombres et compteurs", Description: "Compter l'âge, les objets, les personnes et l'heure.", Level: "A1", Order: 12, Content: `OBJECTIF
Choisir entre nombres coréens natifs et sino-coréens dans les situations courantes.

RÈGLE
Natifs : 하나, 둘, 셋, 넷, 다섯. Ils servent notamment avec 개 objets, 명 personnes et 시 heures. Devant un compteur : 한 개, 두 명, 세 시, 네 잔.

Sino-coréens : 일, 이, 삼, 사, 오. Ils servent pour les dates, minutes, prix, numéros et mesures.

EXEMPLES
사과 두 개 two pommes · 학생 세 명 trois étudiants · 커피 한 잔 une tasse de café.
오천 원 5 000 wons · 삼월 이일 le 2 mars · 십 분 dix minutes.

À RETENIR
하나, 둘, 셋, 넷 deviennent 한, 두, 세, 네 devant un compteur.

PRATIQUE
Exprime : trois cafés, quatre personnes, 7 h 20 et 10 000 wons.

CORRIGÉ
커피 세 잔, 네 명, 일곱 시 이십 분, 만 원.`},
		{ID: "daily-sentences", Title: "Raconter sa journée", Description: "Enchaîner temps, lieux, objets et verbes dans un récit simple.", Level: "A1", Order: 13, Content: `OBJECTIF
Produire un petit paragraphe cohérent au présent ou au passé.

MÉTHODE
Commence par le moment, ajoute le lieu avec 에/에서, l'objet avec 을/를, puis termine par le verbe.

EXEMPLES
아침 일곱 시에 일어나요. À sept heures du matin, je me lève.
카페에서 커피를 마셔요. Je bois un café au café.
저녁에 한국어를 공부해요. Le soir, j'étudie le coréen.

MODÈLE
오늘 아침에 학교에 갔어요. 학교에서 한국어를 공부했어요. 저녁에는 친구를 만나서 같이 밥을 먹었어요.

À RETENIR
Le sujet « je » peut rester implicite. Répéter 저는 à chaque phrase paraît peu naturel.

PRATIQUE
Écris quatre phrases sur ta journée avec 오늘, 에/에서, un objet et au moins un verbe au passé.

CORRIGÉ
Exemple : 오늘 아침에 집에서 커피를 마셨어요. 회사에 갔어요. 저녁에 친구를 만났어요. 집에서 한국어를 공부했어요.`},
		{ID: "grammar-connectors", Title: "Relier ses idées", Description: "Utiliser 그리고, 하지만, 그래서 et -고.", Level: "A1", Order: 14, Content: `OBJECTIF
Passer de phrases isolées à un discours court et naturel.

RÈGLE
그리고 ajoute une idée, 하지만 introduit un contraste, 그래서 présente une conséquence. La terminaison -고 relie deux verbes ou adjectifs : 먹고 마셔요.

EXEMPLES
한국어를 공부해요. 그리고 한국 음악을 들어요.
커피를 좋아해요. 하지만 차는 안 좋아해요.
오늘 바빠요. 그래서 못 가요.
친구를 만나고 같이 밥을 먹어요.

À RETENIR
-고 relie simplement. Pour exprimer clairement une cause, utilise d'abord 그래서 entre deux phrases.

PRATIQUE
Relie : « J'étudie le coréen. J'écoute de la musique coréenne. C'est difficile. C'est intéressant. »

CORRIGÉ
한국어를 공부하고 한국 음악을 들어요. 어려워요. 하지만 재미있어요.`},
	}
}
