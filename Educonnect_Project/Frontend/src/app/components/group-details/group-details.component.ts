import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { HttpClientModule } from '@angular/common/http';
import { RouterModule } from '@angular/router';


@Component({
  selector: 'app-group-details',
  standalone: true,
  imports: [CommonModule,HttpClientModule,RouterModule],
  templateUrl: './group-details.component.html',
  styleUrls: ['./group-details.component.scss']
})
export class GroupDetailsComponent implements OnInit {
  groupId!: number;
  group: any;
  members: any[] = [];
  token = localStorage.getItem("token");

  constructor(private route: ActivatedRoute, private http: HttpClient) {}

  ngOnInit(): void {
    this.groupId = Number(this.route.snapshot.paramMap.get('id'));
    this.loadGroupDetails();
    this.loadMembers();
  }

  getAuthHeaders() {
    return new HttpHeaders({
      'Authorization': `Bearer ${this.token}`
    });
  }

  loadGroupDetails() {
    this.http.get(`http://localhost:8080/groups/${this.groupId}`, {
      headers: this.getAuthHeaders()
    }).subscribe(data => this.group = data);
  }

  loadMembers() {
    this.http.get<any[]>(`http://localhost:8080/groups/${this.groupId}/members`, {
      headers: this.getAuthHeaders()
    }).subscribe(data => this.members = data);
  }
}
