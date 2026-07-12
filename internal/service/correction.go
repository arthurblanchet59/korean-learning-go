package service

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

var commonKoreanCorrections = []struct {
	from   string
	to     string
	reason string
}{
	{"안녕 하세요", "안녕하세요", "안녕하세요 s'ecrit en un seul mot."},
	{"감사 합니다", "감사합니다", "La terminaison 합니다 est attachee au radical."},
	{"죄송 합니다", "죄송합니다", "La terminaison 합니다 est attachee au radical."},
	{"공부 해요", "공부해요", "하다 et sa terminaison se contractent en 해요."},
	{"좋아 해요", "좋아해요", "좋아하다 se conjugue sans espace avant 해요."},
	{"저 는", "저는", "Une particule s'attache au nom ou au pronom."},
	{"나 는", "나는", "Une particule s'attache au nom ou au pronom."},
}

var koreanTokenPattern = regexp.MustCompile(`[가-힣]+(?:은|는|이|가|을|를|이에요|예요)`)

func CorrectKorean(input string) (string, []core.Correction) {
	corrected := strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
	corrections := make([]core.Correction, 0)

	for _, rule := range commonKoreanCorrections {
		if strings.Contains(corrected, rule.from) {
			corrected = strings.ReplaceAll(corrected, rule.from, rule.to)
			corrections = append(corrections, core.Correction{Original: rule.from, Replacement: rule.to, Reason: rule.reason})
		}
	}

	corrected = koreanTokenPattern.ReplaceAllStringFunc(corrected, func(token string) string {
		updated, reason := correctKoreanEnding(token)
		if updated != token {
			corrections = append(corrections, core.Correction{Original: token, Replacement: updated, Reason: reason})
		}
		return updated
	})

	if corrected != "" && !strings.ContainsAny(string([]rune(corrected)[len([]rune(corrected))-1]), ".!?。！？") {
		corrected += "."
		corrections = append(corrections, core.Correction{Original: "fin de phrase", Replacement: ".", Reason: "Ajout d'une ponctuation finale."})
	}
	if containsLatin(corrected) {
		corrections = append(corrections, core.Correction{Original: "alphabet latin", Replacement: "hangeul", Reason: "Essaie de remplacer les mots en alphabet latin par du hangeul."})
	}
	return corrected, corrections
}

func correctKoreanEnding(token string) (string, string) {
	pairs := []struct {
		withBatchim    string
		withoutBatchim string
		reason         string
	}{
		{"이에요", "예요", "이에요 suit une consonne finale; 예요 suit une voyelle."},
		{"은", "는", "은 suit une consonne finale; 는 suit une voyelle."},
		{"이", "가", "이 suit une consonne finale; 가 suit une voyelle."},
		{"을", "를", "을 suit une consonne finale; 를 suit une voyelle."},
	}
	for _, pair := range pairs {
		for _, ending := range []string{pair.withBatchim, pair.withoutBatchim} {
			if !strings.HasSuffix(token, ending) {
				continue
			}
			stem := strings.TrimSuffix(token, ending)
			if stem == "" {
				return token, ""
			}
			expected := pair.withoutBatchim
			if hasBatchim(lastRune(stem)) {
				expected = pair.withBatchim
			}
			return stem + expected, pair.reason
		}
	}
	return token, ""
}

func hasBatchim(r rune) bool {
	return r >= 0xAC00 && r <= 0xD7A3 && (r-0xAC00)%28 != 0
}

func lastRune(value string) rune {
	var last rune
	for _, r := range value {
		last = r
	}
	return last
}

func containsLatin(value string) bool {
	for _, r := range value {
		if unicode.In(r, unicode.Latin) {
			return true
		}
	}
	return false
}
