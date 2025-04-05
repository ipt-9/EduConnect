export function getUserIdFromToken(token: string | null): number | null {
  if (!token) return null;

  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return decoded.user_id ?? null;
  } catch (err) {
    console.error("❌ Fehler beim Dekodieren des Tokens:", err);
    return null;
  }
}
