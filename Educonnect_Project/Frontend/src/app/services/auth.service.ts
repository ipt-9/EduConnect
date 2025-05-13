import { Injectable } from '@angular/core';
import { Router } from '@angular/router';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private checkInterval: any;

  constructor(private router: Router) {
    this.startTokenWatcher();
    this.redirectIfAuthenticatedOnRoot();
  }

  // Holt den gespeicherten Token
  getToken(): string | null {
    return localStorage.getItem('token');
  }

  // Prüft, ob der Token abgelaufen ist
  isTokenExpired(): boolean {
    const token = this.getToken();
    if (!token) return true;

    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Math.floor(Date.now() / 1000);  // Aktuelle Zeit in Sekunden
      return payload.exp < now;
    } catch (err) {
      console.error('❌ Ungültiges Token:', err);
      return true;
    }
  }

  // Prüft, ob der Nutzer sich auf einer Auth-Route befindet
  isAuthRoute(): boolean {
    const currentUrl = this.router.url;
    return currentUrl === '/login' || currentUrl === '/register';
  }

  // Entfernt den Token und leitet zum Login weiter
  logout() {
    console.warn('🚪 Ausgeloggt – Token ist abgelaufen!');
    localStorage.removeItem('token');
    this.router.navigate(['/login']);
  }

  // Startet die Überwachung des Token-Zustands
  startTokenWatcher(intervalMs = 5000) {
    this.checkInterval = setInterval(() => {
      if (this.isTokenExpired()) {
        console.warn('🚪 Token abgelaufen, Nutzer wird ausgeloggt');
        this.logout();  // Immer ausloggen, egal auf welcher Route
      }
    }, intervalMs);
  }

  // Stoppt die Token-Überwachung
  stopTokenWatcher() {
    clearInterval(this.checkInterval);
  }

  // Leitet vom Root weiter, wenn Nutzer eingeloggt ist
  private redirectIfAuthenticatedOnRoot() {
    if (this.router.url === '/' && !this.isTokenExpired()) {
      console.log('🔄 Gültiger Token gefunden – Weiterleitung zu /dashboard');
      this.router.navigate(['/dashboard']);
    }
  }
}
