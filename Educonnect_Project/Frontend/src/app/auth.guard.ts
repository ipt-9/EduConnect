import { Injectable } from '@angular/core';
import { CanActivate, Router } from '@angular/router';

@Injectable({
  providedIn: 'root'
})
export class AuthGuard implements CanActivate {
  constructor(private router: Router) {}

  canActivate(): boolean {
    const token = localStorage.getItem('token');

    // ✅ Token vorhanden?
    if (!token) {
      this.router.navigate(['/login']);
      return false;
    }

    // 🧠 Ablaufzeit prüfen (JWT Payload)
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Math.floor(Date.now() / 1000000); // aktuelle Zeit in Sekunden

      if (payload.exp && payload.exp < now) {
        console.warn('⏳ Token ist abgelaufen!');
        localStorage.removeItem('token');
        this.router.navigate(['/login']);
        return false;
      }

      return true; // ✅ Zugriff erlaubt

    } catch (err) {
      console.error('❌ Ungültiges Token:', err);
      localStorage.removeItem('token');
      this.router.navigate(['/login']);
      return false;
    }
  }
}
