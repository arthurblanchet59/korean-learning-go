import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  answerReviewCard,
  checkCardAnswer,
  fetchCards,
  fetchDecks,
  fetchDifficultCards,
  fetchDueCards,
  fetchJournal,
  fetchLessons,
  fetchStats
} from "../../../shared/api/studyApi.js";

const emptyStats = {
  totalCards: 0,
  dueCards: 0,
  newCards: 0,
  difficultCards: 0,
  masteredCards: 0,
  reviewsToday: 0,
  accuracyPercent: 0,
  currentStreak: 0,
  longestStreak: 0,
  reviewHistory: []
};

export function useStudyDashboard(authToken) {
  const [data, setData] = useState({
    cards: [],
    dueCards: [],
    difficultCards: [],
    decks: [],
    lessons: [],
    journal: [],
    stats: emptyStats
  });
  const [activeIndex, setActiveIndex] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
	const [isMutating, setIsMutating] = useState(false);
	const [apiOnline, setApiOnline] = useState(true);
	const mutationLock = useRef(false);
  const [error, setError] = useState("");

  const reload = useCallback(async () => {
    setIsLoading(true);
    const results = await Promise.all([
      fetchStats(authToken),
      fetchDueCards(authToken),
      fetchCards(authToken),
      fetchDifficultCards(authToken),
      fetchDecks(authToken),
      fetchLessons(authToken),
      fetchJournal(authToken)
    ]);
    const failed = results.find((result) => !result.ok);
	setApiOnline(results.every((result) => result.fromAPI !== false));
    if (failed) {
      setError(failed.error || "Impossible de charger les donnees.");
    } else {
      setError("");
    }
    setData({
      stats: results[0].data ?? emptyStats,
      dueCards: results[1].data ?? [],
      cards: results[2].data ?? [],
      difficultCards: results[3].data ?? [],
      decks: results[4].data ?? [],
      lessons: results[5].data ?? [],
      journal: results[6].data ?? []
    });
    setActiveIndex(0);
    setIsLoading(false);
  }, [authToken]);

  useEffect(() => {
    reload();
  }, [reload]);

  const activeCard = useMemo(() => data.dueCards[activeIndex], [data.dueCards, activeIndex]);

  const answerCard = useCallback(async (rating) => {
	if (!activeCard || mutationLock.current) return false;
	mutationLock.current = true;
	setIsMutating(true);
	try {
	  const result = await answerReviewCard(activeCard.id, rating, authToken);
	  setApiOnline(result.fromAPI !== false);
	  if (!result.ok) {
	    setError(result.error);
	    return false;
	  }
	  await reload();
	  return true;
	} finally {
	  mutationLock.current = false;
	  setIsMutating(false);
    }
  }, [activeCard, authToken, reload]);

  const checkAnswer = useCallback(async (id, answer, direction) => {
    const result = await checkCardAnswer(id, answer, direction, authToken);
    if (!result.ok) {
      setError(result.error || "Impossible de verifier la reponse.");
      return null;
    }
    setError("");
    return result.data;
  }, [authToken]);

  const runMutation = useCallback(async (operation) => {
	if (mutationLock.current) return { ok: false, error: "Une opération est déjà en cours." };
	mutationLock.current = true;
	setIsMutating(true);
	let result;
	try {
	  result = await operation();
	} finally {
	  mutationLock.current = false;
	  setIsMutating(false);
	}
	setApiOnline(result.fromAPI !== false);
    if (!result.ok) {
      setError(result.error || "L'operation a echoue.");
      return result;
    }
    setError("");
    await reload();
    return result;
  }, [reload]);

  return {
    ...data,
    activeCard,
    activeIndex,
	apiOnline,
    error,
	isLoading,
	isMutating,
    answerCard,
    checkAnswer,
    reload,
    runMutation,
    selectCard: setActiveIndex
  };
}
