import { Component, EventEmitter, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { Router } from '@angular/router';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    FormsModule,
    HttpClientModule,
  ],
  templateUrl: './sidebar.component.html',
  styleUrls: ['./sidebar.component.scss']
})
export class SidebarComponent {
  isExpanded = true;

  @Output() expandedChange = new EventEmitter<boolean>();

  constructor(private http: HttpClient, private router: Router) {} // ✅ INSIDE the class
  showComingSoon = false;

  showComingSoonPopup(): void {
    this.showComingSoon = true;
    setTimeout(() => this.showComingSoon = false, 3000);
  }

  toggleSidebar(): void {
    this.isExpanded = !this.isExpanded;
    this.expandedChange.emit(this.isExpanded);
  }
  logout() {
    this.http.post('https://api.educonnect-bmsd22a.bbzwinf.ch/logout', {}).subscribe({
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

  navigateTo(path: string): void {
    this.router.navigate([path]);
  }
}
