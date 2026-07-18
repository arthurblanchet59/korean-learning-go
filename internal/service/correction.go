package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

var ErrCorrectionUnavailable = errors.New("automatic correction is unavailable")
var ErrEmbeddingUnavailable = errors.New("pedagogical search is unavailable")

type KoreanCorrector interface {
	Correct(ctx context.Context, input string) (core.CorrectionResult, error)
}

type LocalKoreanCorrector struct{}

func (LocalKoreanCorrector) Correct(_ context.Context, input string) (core.CorrectionResult, error) {
	corrected, corrections := CorrectKorean(input)
	return core.CorrectionResult{
		CorrectedText: corrected,
		Corrections:   corrections,
		Sources:       []core.CorrectionSource{},
	}, nil
}

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

func CorrectKorean(input string) (string, []core.Correction) {
	corrected := strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
	corrections := make([]core.Correction, 0)

	for _, rule := range commonKoreanCorrections {
		if strings.Contains(corrected, rule.from) {
			corrected = strings.ReplaceAll(corrected, rule.from, rule.to)
			corrections = append(corrections, core.Correction{Original: rule.from, Replacement: rule.to, Reason: rule.reason})
		}
	}

	if corrected != "" && !strings.ContainsAny(string([]rune(corrected)[len([]rune(corrected))-1]), ".!?。！？") {
		corrected += "."
		corrections = append(corrections, core.Correction{Original: "fin de phrase", Replacement: ".", Reason: "Ajout d'une ponctuation finale."})
	}
	if containsLatin(corrected) {
		corrections = append(corrections, core.Correction{Original: "alphabet latin", Replacement: "hangeul", Reason: "Essaie de remplacer les mots en alphabet latin par du hangeul."})
	}
	return corrected, corrections
}

func containsLatin(value string) bool {
	for _, r := range value {
		if unicode.In(r, unicode.Latin) {
			return true
		}
	}
	return false
}

func containsHangul(value string) bool {
	for _, r := range value {
		if unicode.In(r, unicode.Hangul) {
			return true
		}
	}
	return false
}
