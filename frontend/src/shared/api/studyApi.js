import { fallbackCards, fallbackDecks, fallbackStats } from "../data/fallbackStudyData.js";
import { getJSON, postJSON } from "./client.js";

export function fetchStats(token) {
  return getJSON("/stats", fallbackStats, token);
}

export function fetchDueCards(token) {
  return getJSON("/reviews/due", fallbackCards, token);
}

export function fetchDecks(token) {
  return getJSON("/decks", fallbackDecks, token);
}

export function answerReviewCard(cardID, rating, token) {
  return postJSON(`/reviews/${cardID}/answer`, { rating }, token);
}
