package service

import (
	"context"
	"testing"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type correctionStub struct{}

func (correctionStub) Correct(_ context.Context, input string) (core.CorrectionResult, error) {
	return core.CorrectionResult{
		CorrectedText: input + " corrigé",
		Corrections:   []core.Correction{{Original: input, Replacement: input + " corrigé", Reason: "Test"}},
		Sources:       []core.CorrectionSource{},
	}, nil
}

func TestCorrectKoreanSpacingAndParticles(t *testing.T) {
	corrected, corrections := CorrectKorean("저 는 오늘 한국어를 공부 해요")
	if corrected != "저는 오늘 한국어를 공부해요." {
		t.Fatalf("unexpected correction: %q", corrected)
	}
	if len(corrections) != 3 {
		t.Fatalf("expected 3 explained corrections, got %d", len(corrections))
	}
}

func TestCorrectKoreanDoesNotRewriteAmbiguousWordEndings(t *testing.T) {
	for _, input := range []string{"아이 예뻐요", "마을 좋아요"} {
		corrected, _ := CorrectKorean(input)
		if corrected != input+"." {
			t.Fatalf("valid Korean was rewritten: input=%q corrected=%q", input, corrected)
		}
	}
}

func TestNormalizeAnswerIgnoresFrenchDiacritics(t *testing.T) {
	if normalizeAnswer("À l'école !") != normalizeAnswer("a lecole") {
		t.Fatal("expected French diacritics and punctuation to be ignored")
	}
}

func TestStudyServiceUsesInjectedKoreanCorrector(t *testing.T) {
	study := &StudyService{corrector: correctionStub{}}
	result, err := study.CorrectJournalText(context.Background(), "문장")
	if err != nil {
		t.Fatal(err)
	}
	if result.CorrectedText != "문장 corrigé" || len(result.Corrections) != 1 {
		t.Fatalf("unexpected injected correction: %+v", result)
	}
}
