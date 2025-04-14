import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { HttpClientModule } from '@angular/common/http';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-group-details',
  standalone: true,
  imports: [CommonModule, HttpClientModule, RouterModule],
  templateUrl: './group-details.component.html',
  styleUrls: ['./group-details.component.scss']
})
export class GroupDetailsComponent implements OnInit {
  groupId!: number;
  group: any;
  members: any[] = [];
  notifications: any[] = [];

  token = localStorage.getItem("token");
  userId!: number;
  isCurrentUserAdmin: boolean = false;

  constructor(
    private route: ActivatedRoute,
    private http: HttpClient,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.groupId = Number(this.route.snapshot.paramMap.get('id'));

    const decoded = this.decodeToken();
    this.userId = decoded?.user_id;

    this.loadGroupDetails();
    this.loadMembers();
    this.loadNotifications();
  }

  getAuthHeaders() {
    return new HttpHeaders({
      'Authorization': `Bearer ${this.token}`
    });
  }

  decodeToken(): any {
    if (!this.token) return null;
    try {
      const payload = this.token.split('.')[1];
      const decoded = atob(payload);
      return JSON.parse(decoded);
    } catch (e) {
      console.error("❌ Fehler beim Dekodieren des Tokens:", e);
      return null;
    }
  }

  loadGroupDetails() {
    this.http.get(`http://localhost:8080/groups/${this.groupId}`, {
      headers: this.getAuthHeaders()
    }).subscribe(data => this.group = data);
  }

  loadMembers() {
    this.http.get<any[]>(`http://localhost:8080/groups/${this.groupId}/members`, {
      headers: this.getAuthHeaders()
    }).subscribe(data => {
      this.members = data;
      const me = data.find(m => m.user_id === this.userId);
      this.isCurrentUserAdmin = me?.role === 'admin';
    });
  }

  loadNotifications() {
    this.http.get<any[]>(`http://localhost:8080/groups/${this.groupId}/notifications`, {
      headers: this.getAuthHeaders()
    }).subscribe({
      next: data => this.notifications = data,
      error: err => console.error("Fehler beim Laden der Notifications:", err)
    });
  }

  leaveGroup() {
    if (!this.userId) {
      alert("User-ID nicht gefunden");
      return;
    }

    const confirmLeave = confirm("Möchtest du die Gruppe wirklich verlassen?");
    if (!confirmLeave) return;

    this.http.delete(`http://localhost:8080/groups/${this.groupId}/members/${this.userId}`, {
      headers: this.getAuthHeaders()
    }).subscribe({
      next: () => {
        alert("✅ Du hast die Gruppe erfolgreich verlassen.");
        this.router.navigate(['/groupmanager']);
      },
      error: err => {
        if (err.status === 403 && err.error?.includes("Admins können sich nicht selbst entfernen")) {
          alert("⚠️ Du bist der Admin dieser Gruppe und kannst sie nicht verlassen. Übertrage zuerst die Admin-Rolle an ein anderes Mitglied.");
        } else {
          alert("❌ Fehler beim Verlassen der Gruppe.");
        }
        console.error("Fehler beim Verlassen der Gruppe:", err);
      }
    });
  }
  getNotificationClass(message: string): string {
    if (message.includes('hat die Aufgabe')) {
      return 'notif-success';
    } else if (message.includes('Gruppe verlassen')) {
      return 'notif-danger';
    } else if (message.includes('Gruppe beigetreten')) {
      return 'notif-success';
    } else if (message.includes('zum Gruppen-')) {
      return 'notif-warning';
    } else {
      return 'notif-default';
    }
  }

}
