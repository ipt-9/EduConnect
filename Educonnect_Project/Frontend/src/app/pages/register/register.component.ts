import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpParams, HttpClientModule } from '@angular/common/http';
import { Router } from '@angular/router';

@Component({
  standalone: true,
  selector: 'app-register',
  imports: [FormsModule, CommonModule, HttpClientModule],
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.scss']
})
export class RegisterComponent {
  username: string = '';
  email: string = '';
  password: string = '';

  loading: boolean = false;
  successMessage: string = '';

  constructor(private http: HttpClient, private router: Router) {}

  LoginBtn () {
    this.router.navigate(['/login']);
  }

  RegisterBtn () {
    this.router.navigate(['/register']);
  }
  register() {
    this.loading = true;
    this.successMessage = '';

    const body = new HttpParams()
      .set('username', this.username)
      .set('email', this.email)
      .set('password', this.password);

    this.http.post('http://localhost:8080/register', body.toString(), {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      responseType: 'text'
    }).subscribe({
      next: () => {
        this.loading = false;
        this.successMessage = '✅ Benutzer erfolgreich erstellt!';

        // Weiterleitung nach kurzer Pause
        setTimeout(() => {
          this.router.navigate(['/login']);
        }, 1000);
      },
      error: err => {
        this.loading = false;
        this.successMessage = '❌ Fehler bei der Registrierung.';
        console.error('Fehler bei der Registrierung:', err);
      }
    });
  }
}
