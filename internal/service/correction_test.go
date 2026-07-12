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

func TestCorrectKoreanBatchimParticle(t *testing.T) {
	corrected, _ := CorrectKorean("학교은 좋아요")
	if corrected != "학교는 좋아요." {
		t.Fatalf("expected topic particle correction, got %q", corrected)
	}
}

func TestNormalizeAnswerIgnoresFrenchDiacritics(t *testing.T) {
	if normalizeAnswer("À l'école !") != normalizeAnswer("a lecole") {
		t.Fatal("expected French diacritics and punctuation to be ignored")
	}
}
