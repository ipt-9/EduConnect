import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { Router } from '@angular/router';

@Component({
  standalone: true,
  selector: 'app-login',
  imports: [FormsModule, CommonModule, HttpClientModule],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {
  email = '';
  password = '';
  code = '';

  loading = false;
  success = false;
  errorMessage = '';
  verifyError = '';
  show2FA = false;

  constructor(private http: HttpClient, private router: Router) {}

  LoginBtn () {
    this.router.navigate(['/login']);
  }

  RegisterBtn () {
    this.router.navigate(['/register']);
  }
  login() {
    this.resetUI();
    this.loading = true;

    this.http.post('http://localhost:8080/login', {
      email: this.email,
      password: this.password
    }, {
      responseType: 'text'
    }).subscribe({
      next: () => {
        this.loading = false;
        this.show2FA = true; // Öffne 2FA-Popup
      },
      error: err => {
        this.loading = false;
        this.errorMessage = '❌ Login fehlgeschlagen. Bitte überprüfen.';
        console.error(err);
      }
    });
  }
  showSuccessAnimation = false;

  verify2FA()   {
    this.verifyError = '';

    if (!this.code || this.code.length !== 6) {
      this.verifyError = '⚠️ Bitte gib einen 6-stelligen Code ein.';
      return;
    }

    this.http.post<{ token: string }>('http://localhost:8080/verify-2fa', {
      email: this.email,
      code: this.code
    }).subscribe({
      next: (res) => {
        localStorage.setItem('token', res.token); // Token speichern
        this.success = true;
        this.show2FA = false;
        this.showSuccessAnimation = true;

        setTimeout(() => {
          this.showSuccessAnimation = false;
          this.router.navigate(['/homepage']);
        }, 2000);

      },
      error: err => {
        this.verifyError = '❌ Ungültiger oder abgelaufener Code.';
        console.error(err);
      }
    });
  }

  cancel2FA() {
    this.show2FA = false;
    this.code = '';
    this.verifyError = '';
  }

  private resetUI() {
    this.errorMessage = '';
    this.verifyError = '';
    this.success = false;
    this.code = '';
    this.show2FA = false;
  }
}
