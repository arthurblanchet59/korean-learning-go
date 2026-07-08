import { fallbackCards, fallbackDecks, fallbackStats } from "../data/fallbackStudyData.js";
import { getJSON, postJSON } from "./client.js";

export function fetchStats() {
  return getJSON("/stats", fallbackStats);
}

export function fetchDueCards() {
  return getJSON("/reviews/due", fallbackCards);
}

export function fetchDecks() {
  return getJSON("/decks", fallbackDecks);
}

export function answerReviewCard(cardID, rating) {
  return postJSON(`/reviews/${cardID}/answer`, { rating });
}
