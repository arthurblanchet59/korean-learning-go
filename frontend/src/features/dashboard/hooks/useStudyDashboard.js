import { useCallback, useEffect, useMemo, useState } from "react";

import { answerReviewCard, fetchDecks, fetchDueCards, fetchStats } from "../../../shared/api/studyApi.js";
import { fallbackCards, fallbackDecks, fallbackStats } from "../../../shared/data/fallbackStudyData.js";

export function useStudyDashboard(authToken) {
  const [cards, setCards] = useState([]);
  const [decks, setDecks] = useState([]);
  const [stats, setStats] = useState(fallbackStats);
  const [activeIndex, setActiveIndex] = useState(0);
  const [apiOnline, setAPIOnline] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const loadDashboard = useCallback(async () => {
    setIsLoading(true);

    const [statsResult, cardsResult, decksResult] = await Promise.all([
      fetchStats(authToken),
      fetchDueCards(authToken),
      fetchDecks(authToken)
    ]);

    setAPIOnline(statsResult.fromAPI || cardsResult.fromAPI || decksResult.fromAPI);
    setStats(statsResult.data ?? fallbackStats);
    setCards(cardsResult.data ?? fallbackCards);
    setDecks(decksResult.data ?? fallbackDecks);
    setActiveIndex(0);
    setIsLoading(false);
  }, [authToken]);

  useEffect(() => {
    loadDashboard();
  }, [loadDashboard]);

  const activeCard = useMemo(() => cards[activeIndex], [cards, activeIndex]);

  const refreshStats = useCallback(async () => {
    const result = await fetchStats(authToken);
    setStats(result.data ?? fallbackStats);
  }, [authToken]);

  const answerCard = useCallback(
    async (rating) => {
      const card = cards[activeIndex];
      if (!card) {
        return;
      }

      if (!apiOnline) {
        setActiveIndex((current) => (cards.length === 0 ? 0 : (current + 1) % cards.length));
        return;
      }

      const result = await answerReviewCard(card.id, rating, authToken);
      if (!result.ok) {
        return;
      }

      setCards((currentCards) => {
        const nextCards = currentCards.filter((currentCard) => currentCard.id !== card.id);
        setActiveIndex((current) => Math.min(current, Math.max(nextCards.length - 1, 0)));
        return nextCards;
      });
      await refreshStats();
    },
    [activeIndex, apiOnline, authToken, cards, refreshStats]
  );

  const selectCard = useCallback((index) => {
    setActiveIndex(index);
  }, []);

  const restartSession = useCallback(() => {
    setActiveIndex(0);
  }, []);

  return {
    apiOnline,
    cards,
    decks,
    stats,
    activeCard,
    activeIndex,
    isLoading,
    answerCard,
    selectCard,
    restartSession
  };
}
