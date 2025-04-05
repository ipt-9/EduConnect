import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { AuthService } from './services/auth.service';
import { LogoutButtonComponent } from './components/logout-button/logout-button.component'; // 👈 wichtig!
import { CommonModule } from '@angular/common'; // 👈 wichtig!
import { HttpClient, HttpClientModule } from '@angular/common/http';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet,LogoutButtonComponent,CommonModule,HttpClientModule],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
  title = 'Frontend';

  constructor(public authService: AuthService) {
    // 🔒 Token-Überwachung startet automatisch im AuthService
  }
}
