import { deleteJSON, getJSON, getText, postJSON, putJSON } from "./client.js";

export const fetchStats = (token) => getJSON("/stats", null, token);
export const fetchDueCards = (token) => getJSON("/reviews/due", [], token);
export const fetchCards = (token) => getJSON("/cards", [], token);
export const fetchDifficultCards = (token) => getJSON("/cards/difficult", [], token);
export const fetchDecks = (token) => getJSON("/decks", [], token);
export const fetchLessons = (token) => getJSON("/lessons", [], token);
export const fetchJournal = (token) => getJSON("/journal", [], token);

export const answerReviewCard = (id, rating, token) => postJSON(`/reviews/${id}/answer`, { rating }, token);
export const checkCardAnswer = (id, answer, direction, token) => postJSON(`/study/cards/${id}/check`, { answer, direction }, token);

export const createDeck = (payload, token) => postJSON("/decks", payload, token);
export const updateDeck = (id, payload, token) => putJSON(`/decks/${id}`, payload, token);
export const deleteDeck = (id, token) => deleteJSON(`/decks/${id}`, undefined, token);
export const bulkUpdateDecks = (ids, patch, token) => putJSON("/decks/bulk", { ids, patch }, token);
export const bulkDeleteDecks = (ids, token) => deleteJSON("/decks/bulk", { ids }, token);

export const createCard = (payload, token) => postJSON("/cards", payload, token);
export const updateCard = (id, payload, token) => putJSON(`/cards/${id}`, payload, token);
export const deleteCard = (id, token) => deleteJSON(`/cards/${id}`, undefined, token);
export const bulkUpdateCards = (ids, patch, token) => putJSON("/cards/bulk", { ids, patch }, token);
export const bulkDeleteCards = (ids, token) => deleteJSON("/cards/bulk", { ids }, token);
export const importCardsCSV = (deckId, csv, token) => postJSON("/cards/import", { deckId, csv }, token);
export const exportCardsCSV = (token) => getText("/cards/export", token);

export const searchAll = (query, token) => getJSON(`/search?query=${encodeURIComponent(query)}`, { decks: [], cards: [] }, token);
export const updateLessonProgress = (id, payload, token) => putJSON(`/lessons/${id}/progress`, payload, token);

export const createJournalEntry = (payload, token) => postJSON("/journal", payload, token);
export const previewJournalCorrection = (payload, token) => postJSON("/journal/correct", payload, token);
export const updateJournalEntry = (id, payload, token) => putJSON(`/journal/${id}`, payload, token);
export const deleteJournalEntry = (id, token) => deleteJSON(`/journal/${id}`, undefined, token);

export const resetDatabase = (token) => postJSON("/admin/reset", {}, token);
export const adminUpdateUser = (id, payload, token) => putJSON(`/admin/users/${id}`, payload, token);
