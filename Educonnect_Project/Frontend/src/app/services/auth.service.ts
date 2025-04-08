import { Injectable } from '@angular/core';
import { Router } from '@angular/router';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private checkInterval: any;

  constructor(private router: Router) {
    this.startTokenWatcher();
  }

  getToken(): string | null {
    return localStorage.getItem('token');
  }

  isTokenExpired(): boolean {
    const token = this.getToken();
    if (!token) return true;

    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Math.floor(Date.now() / 1000); // Zeit in Sekunden
      return payload.exp < now;
    } catch (err) {
      console.error('❌ Ungültiges Token:', err);
      return true;
    }
  }
  isAuthRoute(): boolean {
    const currentUrl = this.router.url;
    return currentUrl === '/login' || currentUrl === '/register';
  }
  logout() {
    console.warn('🚪 Ausgeloggt – Token ist abgelaufen!');
    localStorage.removeItem('token');
    this.router.navigate(['/login']);
  }

  startTokenWatcher(intervalMs = 5000) {
    this.checkInterval = setInterval(() => {
      if (this.isTokenExpired()) {
        const publicRoutes = ['/', '/login', '/register'];

        // Nur logout, wenn man NICHT auf einer öffentlichen Seite ist
        if (!publicRoutes.includes(this.router.url)) {
          this.logout();
        } else {
          // 🔕 Kein Redirect, nur Token entfernen
          console.warn('⏳ Token abgelaufen, aber auf öffentlicher Route → kein Redirect');
          localStorage.removeItem('token');
        }
      }
    }, intervalMs);
  }


  stopTokenWatcher() {
    clearInterval(this.checkInterval);
  }
}
