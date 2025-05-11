import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { HttpClientModule } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { SidebarComponent } from '../sidebar/sidebar.component';

@Component({
  selector: 'app-group-role-manager',
  standalone: true,
  templateUrl: './group-role-manager.component.html',
  styleUrls: ['./group-role-manager.component.scss'],
  imports: [CommonModule, RouterModule, HttpClientModule, SidebarComponent],
})
export class GroupRoleManagerComponent implements OnInit {
  groupId!: number;
  members: any[] = [];
  currentUserId!: number;
  currentUserRole: string = '';
  token = localStorage.getItem("token");

  constructor(
    private http: HttpClient,
    private route: ActivatedRoute,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.groupId = Number(this.route.snapshot.paramMap.get('id'));
    const decoded = this.decodeToken();
    this.currentUserId = decoded?.user_id;

    this.loadMembers();
  }

  getHeaders() {
    return {
      headers: new HttpHeaders({ 'Authorization': `Bearer ${this.token}` }),
    };
  }

  decodeToken(): any {
    if (!this.token) return null;
    try {
      const payload = this.token.split('.')[1];
      return JSON.parse(atob(payload));
    } catch (e) {
      console.error('❌ Fehler beim JWT-Decode:', e);
      return null;
    }
  }

  loadMembers() {
    this.http.get<any[]>(`http://localhost:8080/groups/${this.groupId}/members`, this.getHeaders())
      .subscribe(data => {
        this.members = data;

        // ✅ Wichtig: user_id vergleichen, nicht id aus group_members
        const me = data.find(m => m.user_id === this.currentUserId);
        this.currentUserRole = me?.role || '';

        if (this.currentUserRole !== 'admin') {
          alert('⛔ Zugriff nur für Admins erlaubt!');
          this.router.navigate(['/']);
        }
      });
  }

  changeRole(member: any, newRole: string) {
    if (!confirm(`Rolle von ${member.username} zu ${newRole} ändern?`)) return;

    this.http.put<{ message: string }>(
      `http://localhost:8080/groups/${this.groupId}/members/${member.user_id}/role`,
      { role: newRole },
      this.getHeaders()
    ).subscribe({
      next: (res) => {
        alert(res.message || '✅ Rolle aktualisiert');
        this.loadMembers(); // Reload
      },
      error: err => {
        console.error('Fehler beim Aktualisieren:', err);
        alert('❌ Fehler beim Aktualisieren der Rolle');
      }
    });
  }
}
