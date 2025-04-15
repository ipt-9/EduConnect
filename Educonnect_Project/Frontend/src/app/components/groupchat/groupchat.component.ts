import {
  Component,
  OnInit,
  OnDestroy,
  ViewChild,
  ElementRef
} from '@angular/core';
import { WebSocketSubject } from 'rxjs/webSocket';
import { ActivatedRoute, Router } from '@angular/router';
import { HttpClient, HttpHeaders, HttpClientModule } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

interface GroupChatMessage {
  linked_task_id?: number | null; // ✅ optional + korrekt typisiert
  message: string;
  created_at: string;
  message_type?: string;
  user: {
    id: number;
    username: string;
    email: string;
    profile_picture_url?: string;
  };
}

function getUserPayloadFromToken(token: string): any {
  try {
    const payload = token.split('.')[1];
    return JSON.parse(atob(payload));
  } catch (err) {
    console.error("Fehler beim Dekodieren des Tokens:", err);
    return {};
  }
}

@Component({
  standalone: true,
  selector: 'app-group-chat',
  templateUrl: './groupchat.component.html',
  styleUrls: ['./groupchat.component.scss'],
  imports: [CommonModule, FormsModule, HttpClientModule]
})
export class GroupChatComponent implements OnInit, OnDestroy {
  groupId!: number;
  token: string | null = '';
  currentUserId = 0;
  currentUsername = '';
  currentEmail = '';
  profilePictureUrl?: string;

  messages: GroupChatMessage[] = [];
  messageText = '';
  private socket!: WebSocketSubject<any>;

  submissions: { task_id: number; task_title: string }[] = [];
  showSubmissionList = false;

  @ViewChild('scrollContainer') scrollContainer!: ElementRef;

  constructor(
    private route: ActivatedRoute,
    private http: HttpClient,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.groupId = Number(this.route.snapshot.paramMap.get('id'));
    this.token = localStorage.getItem('token');

    if (!this.groupId || !this.token) {
      console.error('Token oder Gruppen-ID fehlen');
      return;
    }

    const decoded = getUserPayloadFromToken(this.token);
    this.currentUserId = decoded.user_id || 0;
    this.currentUsername = decoded.username || decoded.email || 'Du';
    this.currentEmail = decoded.email || '';
    this.profilePictureUrl = decoded.profile_picture_url;

    this.loadHistory();
    this.connectWebSocket();
  }

  ngOnDestroy(): void {
    this.socket?.complete();
  }

  extractTaskTitle(message: string): string {
    const match = message.match(/Aufgabe „(.+?)“/);
    return match ? match[1] : 'Unbekannt';
  }

  extractExecutionTime(message: string): string {
    const match = message.match(/🕒 (\d+)ms/);
    return match ? match[1] : '-';
  }

  extractTaskId(message: string): number | null {
    const match = message.match(/taskId=(\d+)/);
    return match ? parseInt(match[1], 10) : null;
  }

  openSubmission(msg: GroupChatMessage): void {
    if (!msg.linked_task_id) {
      console.warn("⚠️ Keine taskId in Nachricht vorhanden:", msg);
      return;
    }

    console.log("➡️ Navigiere zu taskId =", msg.linked_task_id);
    this.router.navigate(['/codingSpace'], {
      queryParams: { taskId: msg.linked_task_id }
    });
  }


  getAuthHeaders() {
    const token = localStorage.getItem('token');
    return new HttpHeaders({
      Authorization: `Bearer ${token ?? ''}`
    });
  }

  loadHistory() {
    this.http.get<any[]>(`http://localhost:8080/groups/${this.groupId}/messages`, {
      headers: this.getAuthHeaders()
    }).subscribe(data => {
      // 🔄 MessageType normalisieren
      this.messages = data.map(m => ({
        ...m,
        message_type: m.MessageType ?? m.message_type ?? 'text'
      })).reverse();
      setTimeout(() => this.scrollToBottom(), 0);
    });
  }


  connectWebSocket() {
    this.socket = new WebSocketSubject(`ws://localhost:8080/ws/groups/${this.groupId}/chat?token=${this.token}`);

    this.socket.subscribe({
      next: (msg: GroupChatMessage) => {
        // 🔄 MessageType normalisieren
        if ((msg as any).MessageType) {
          msg.message_type = (msg as any).MessageType;
        }
        this.messages.push(msg);
        setTimeout(() => this.scrollToBottom(), 0);
      },
      error: err => console.error('WebSocket Fehler:', err),
      complete: () => console.log('WebSocket Verbindung geschlossen')
    });
  }


  sendMessage() {
    if (!this.messageText.trim()) return;

    const newMessage = { message: this.messageText, type: "text" };
    this.socket.next(newMessage);
    this.messageText = '';
  }

  loadSubmissions() {
    this.showSubmissionList = !this.showSubmissionList;
    if (this.submissions.length > 0) return;

    this.http.get<{ task_id: number, task_title: string }[]>(
      `http://localhost:8080/users/me/submissions`,
      { headers: this.getAuthHeaders() }
    ).subscribe({
      next: (res) => this.submissions = res,
      error: (err) => console.error("Fehler beim Laden der Submissions:", err)
    });
  }

  shareSubmission(taskId: number) {
    this.http.post(
      `http://localhost:8080/groups/${this.groupId}/share-submission`,
      { task_id: taskId },
      { headers: this.getAuthHeaders() }
    ).subscribe({
      next: () => console.log("Submission geteilt"),
      error: (err) => console.error("Fehler beim Teilen:", err)
    });
  }

  isOwnMessage(msg: GroupChatMessage): boolean {
    return msg.user.id === this.currentUserId;
  }

  private scrollToBottom(): void {
    try {
      this.scrollContainer.nativeElement.scrollTop = this.scrollContainer.nativeElement.scrollHeight;
    } catch (err) {
      console.error('Scroll error:', err);
    }
  }
}
