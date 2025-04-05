import { Component, OnInit } from '@angular/core';
import { HttpClient, HttpHeaders, HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

@Component({
  standalone: true,
  selector: 'app-group-manager',
  templateUrl: './group-manager.component.html',
  styleUrls: ['./group-manager.component.scss'],
  imports: [CommonModule, FormsModule, HttpClientModule,RouterModule]
})
export class GroupManagerComponent implements OnInit {
  groups: any[] = [];
  groupName = '';
  groupDescription = '';
  joinCode = '';
  token = localStorage.getItem("token");

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    if (!this.token) {
      alert("🔒 Du bist nicht eingeloggt!");
      return;
    }
    this.loadGroups();
  }

  getAuthHeaders(): HttpHeaders {
    return new HttpHeaders({
      'Authorization': `Bearer ${this.token}`
    });
  }

  loadGroups() {
    this.groups = []; // vorher leeren
    this.http.get<any[]>('http://localhost:8080/groups', { headers: this.getAuthHeaders() })
      .subscribe({
        next: (data) => this.groups = data,
        error: (err) => {
          console.error("Fehler beim Laden der Gruppen:", err);
          alert("❌ Gruppen konnten nicht geladen werden.");
        }
      });
  }

  createGroup() {
    if (!this.groupName.trim()) {
      alert("⚠️ Bitte Gruppennamen angeben.");
      return;
    }

    const payload = {
      name: this.groupName,
      description: this.groupDescription
    };

    this.http.post('http://localhost:8080/groups', payload, { headers: this.getAuthHeaders() })
      .subscribe({
        next: () => {
          alert("✅ Gruppe erstellt!");
          this.loadGroups();
          this.groupName = '';
          this.groupDescription = '';
        },
        error: (err) => {
          console.error("Fehler beim Erstellen:", err);
          alert("❌ Gruppe konnte nicht erstellt werden.");
        }
      });
  }

  joinGroup() {
    if (!this.joinCode.trim()) {
      alert("⚠️ Bitte Einladungscode eingeben.");
      return;
    }

    this.http.post(`http://localhost:8080/groups/join?code=${this.joinCode}`, {}, { headers: this.getAuthHeaders() })
      .subscribe({
        next: () => {
          alert("🎉 Erfolgreich beigetreten!");
          this.loadGroups();
          this.joinCode = '';
        },
        error: (err) => {
          console.error("Beitritt fehlgeschlagen:", err);
          alert("❌ Beitritt fehlgeschlagen. Bitte Code prüfen.");
        }
      });
  }

  loadMembers(groupId: number) {
    this.http.get(`http://localhost:8080/groups/${groupId}/members`, { headers: this.getAuthHeaders() })
      .subscribe({
        next: (members) => {
          console.log("👥 Mitglieder:", members);
          alert("Mitglieder:\n" + JSON.stringify(members, null, 2));
        },
        error: (err) => {
          console.error("Fehler beim Laden der Mitglieder:", err);
          alert("❌ Mitglieder konnten nicht geladen werden.");
        }
      });
  }

  updateRole(groupId: number, userId: number, role: string) {
    this.http.put(`http://localhost:8080/groups/${groupId}/members/${userId}/role`, { role }, { headers: this.getAuthHeaders() })
      .subscribe({
        next: () => alert("✅ Rolle aktualisiert."),
        error: (err) => {
          console.error("Fehler beim Rollenwechsel:", err);
          alert("❌ Rolle konnte nicht geändert werden.");
        }
      });
  }
}
