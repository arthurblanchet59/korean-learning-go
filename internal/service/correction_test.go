package service

import "testing"

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
