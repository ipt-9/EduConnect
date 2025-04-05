import {
  Component,
  OnInit,
  OnDestroy,
  ViewChild,
  ElementRef
} from '@angular/core';
import { WebSocketSubject } from 'rxjs/webSocket';
import { ActivatedRoute } from '@angular/router';
import { HttpClient, HttpHeaders, HttpClientModule } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

interface GroupChatMessage {
  message: string;
  created_at: string;
  user: {
    id: number;
    username: string;
    email: string;
    profile_picture_url?: string;
  };
}

// ✅ JWT-Dekodierung ohne externes Paket
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

  @ViewChild('scrollContainer') scrollContainer!: ElementRef;

  constructor(private route: ActivatedRoute, private http: HttpClient) {}

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

  getAuthHeaders() {
    return new HttpHeaders({
      Authorization: `Bearer ${this.token}`
    });
  }

  loadHistory() {
    this.http.get<GroupChatMessage[]>(`http://localhost:8080/groups/${this.groupId}/messages`, {
      headers: this.getAuthHeaders()
    }).subscribe(data => {
      this.messages = data.reverse(); // älteste zuerst
      setTimeout(() => this.scrollToBottom(), 0);
    });
  }

  connectWebSocket() {
    this.socket = new WebSocketSubject(`ws://localhost:8080/ws/groups/${this.groupId}/chat?token=${this.token}`);

    this.socket.subscribe({
      next: (msg: GroupChatMessage) => {
        this.messages.push(msg);
        setTimeout(() => this.scrollToBottom(), 0);
      },
      error: err => console.error('WebSocket Fehler:', err),
      complete: () => console.log('WebSocket Verbindung geschlossen')
    });
  }

  sendMessage() {
    if (!this.messageText.trim()) return;

    const newMessage = { message: this.messageText };
    this.socket.next(newMessage);
    this.messageText = '';
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
