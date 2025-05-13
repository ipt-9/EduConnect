import { Component, OnInit } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Router } from '@angular/router';
import { HttpClientModule } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { SidebarComponent } from '../sidebar/sidebar.component';

@Component({
  selector: 'app-task-list',
  templateUrl: './tasks-list.component.html',
  styleUrls: ['./tasks-list.component.scss'],
  imports: [
    HttpClientModule,CommonModule,SidebarComponent
  ],
})
export class TasksListComponent implements OnInit {
  tasks: any[] = [];

  constructor(private http: HttpClient, private router: Router) {}

  ngOnInit(): void {
    const courseId = localStorage.getItem('activeCourseId'); // vorher beim Klick gesetzt
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<any[]>(`https://api.educonnect-bmsd22a.bbzwinf.ch/courses/${courseId}/tasks`, { headers }).subscribe({
      next: (data) => {
        this.tasks = data;
        console.log('📘 Aufgaben geladen:', data);
      },
      error: (err) => {
        console.error('❌ Fehler beim Laden der Aufgaben:', err);
      }
    });
  }

  openTask(task: any): void {
    localStorage.setItem('activeTask', JSON.stringify(task)); // <--- Das war vorher nicht drin
    localStorage.setItem('activeTaskId', task.id.toString());
    this.router.navigate(['/codingSpace']);
  }

  toggleDescription(task: any): void {
    task.expanded = !task.expanded;
  }
}
