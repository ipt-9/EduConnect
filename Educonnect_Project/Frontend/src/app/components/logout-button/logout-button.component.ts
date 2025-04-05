// src/app/components/logout-button/logout-button.component.ts
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';

@Component({
  standalone: true,
  selector: 'app-logout-button',
  imports: [CommonModule],
  template: `
    <button (click)="logout()" class="logout-btn">🔒 Logout</button>
  `,
  styles: [`
    .logout-btn {
      position: fixed;
      top: 1rem;
      right: 1rem;
      background-color: #ff3b30;
      color: white;
      border: none;
      padding: 0.6rem 1.2rem;
      border-radius: 12px;
      font-weight: bold;
      cursor: pointer;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
      transition: all 0.2s ease-in-out;
    }
    .logout-btn:hover {
      background-color: #e53935;
    }
  `]
})
export class LogoutButtonComponent {
  constructor(private http: HttpClient, private router: Router) {}

  logout() {
    this.http.post('http://localhost:8080/logout', {}).subscribe({
      next: () => {
        localStorage.clear();
        this.router.navigate(['/login']);
      },
      error: () => {
        localStorage.clear();
        this.router.navigate(['/login']);
      }
    });
  }
}
