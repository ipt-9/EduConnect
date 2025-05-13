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
import { SidebarComponent } from '../sidebar/sidebar.component';
import { NgZone } from '@angular/core';


interface GroupChatMessage {
  linked_task_id?: number | null;
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
    console.error("❌ Fehler beim Dekodieren des Tokens:", err);
    return {};
  }
}

@Component({
  standalone: true,
  selector: 'app-group-chat',
  templateUrl: './groupchat.component.html',
  styleUrls: ['./groupchat.component.scss'],
  imports: [CommonModule, FormsModule, HttpClientModule, SidebarComponent]
})
export class GroupChatComponent implements OnInit, OnDestroy {
  groupId!: number;
  token: string | null = '';
  currentUserId = 0;
  currentUsername = '';
  currentEmail = '';
  profilePictureUrl?: string;
  public accessDenied = false;

  messages: GroupChatMessage[] = [];
  messageText = '';
  private socket!: WebSocketSubject<any>;

  submissions: { task_id: number; task_title: string }[] = [];
  showSubmissionList = false;

  @ViewChild('scrollContainer') scrollContainer!: ElementRef;
  @ViewChild('scrollAnchor') scrollAnchor!: ElementRef;


  constructor(
    private route: ActivatedRoute,
    private http: HttpClient,
    private router: Router,
    private ngZone: NgZone
  ) {}

  ngOnInit(): void {
    this.groupId = Number(this.route.snapshot.paramMap.get('id'));
    this.token = localStorage.getItem('token');

    if (!this.groupId || !this.token) {
      console.error('❗ Token oder Gruppen-ID fehlen');
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

  private getAuthHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    return new HttpHeaders({
      Authorization: `Bearer ${token ?? ''}`
    });
  }

  private scrollToBottom(): void {
    this.ngZone.runOutsideAngular(() => {
      requestAnimationFrame(() => {
        try {
          this.scrollAnchor.nativeElement.scrollIntoView({ behavior: 'smooth', block: 'end' });
        } catch (err) {
          console.error('Scroll error:', err);
        }
      });
    });
  }


  loadHistory(): void {
    this.http
      .get<GroupChatMessage[]>(`https://api.educonnect-bmsd22a.bbzwinf.ch/groups/${this.groupId}/messages`, {
        headers: this.getAuthHeaders()
      })
      .subscribe({
        next: data => {
          this.messages = data.map(m => ({
            ...m,
            message_type: m.message_type ?? (m as any).MessageType ?? 'text'
          })).reverse();
          setTimeout(() => this.scrollToBottom(), 0);
        },
        error: err => {
          if (err.status === 403) {
            this.accessDenied = true;  // 🆕 Status aktivieren
          } else {
            console.error('Fehler beim Laden der Nachrichten:', err);
          }
        }
      });
  }


  connectWebSocket(): void {
    this.socket = new WebSocketSubject(
      `https://api.educonnect-bmsd22a.bbzwinf.ch/ws/groups/${this.groupId}/chat?token=${this.token}`
    );

    this.socket.subscribe({
      next: (msg: GroupChatMessage) => {
        msg.message_type = msg.message_type ?? (msg as any).MessageType ?? 'text';
        this.messages.push(msg);
        setTimeout(() => this.scrollToBottom(), 0);
      },
      error: err => console.error('WebSocket Fehler:', err),
      complete: () => console.log('WebSocket Verbindung geschlossen')
    });
  }

  sendMessage(): void {
    if (!this.messageText.trim()) return;

    const newMessage = { message: this.messageText, type: 'text' };
    this.socket.next(newMessage);
    this.messageText = '';
  }

  isOwnMessage(msg: GroupChatMessage): boolean {
    return msg.user.id === this.currentUserId;
  }

  loadSubmissions(): void {
    this.showSubmissionList = !this.showSubmissionList;
    if (this.submissions.length > 0) return;

    this.http
      .get<{ task_id: number; task_title: string }[]>(
        `https://api.educonnect-bmsd22a.bbzwinf.ch/users/me/submissions`,
        { headers: this.getAuthHeaders() }
      )
      .subscribe({
        next: res => (this.submissions = res),
        error: err => console.error('Fehler beim Laden der Submissions:', err)
      });
  }

  shareSubmission(taskId: number): void {
    this.http
      .post(
        `https://api.educonnect-bmsd22a.bbzwinf.ch${this.groupId}/share-submission`,
        { task_id: taskId },
        { headers: this.getAuthHeaders() }
      )
      .subscribe({
        next: () => console.log('✅ Submission geteilt'),
        error: err => console.error('❌ Fehler beim Teilen:', err)
      });
  }

  openSubmission(msg: GroupChatMessage): void {
    if (!msg.linked_task_id) {
      console.warn('⚠️ Keine taskId in Nachricht vorhanden:', msg);
      return;
    }

    this.router.navigate(['/codingSpace'], {
      queryParams: { taskId: msg.linked_task_id }
    });
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
}
