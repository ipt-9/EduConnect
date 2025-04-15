import { Component, OnInit } from '@angular/core';
import { HttpClient, HttpHeaders, HttpClientModule } from '@angular/common/http';
import { CommonModule } from '@angular/common';




@Component({
  selector: 'app-tasks-list',
  standalone: true,
  imports: [HttpClientModule, CommonModule],
  templateUrl: './tasks-list.component.html',
  styleUrls: ['./tasks-list.component.scss']
})
export class TasksListComponent implements OnInit {
  tasks: any[] = [];

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    // 🔁 Dynamisch aus URL holen
    const courseId = localStorage.getItem("activeCourseId");
    if (!courseId) {
      console.error("❌ Keine gültige courseId in der URL gefunden.");
      return;
    }

    this.http.get<any[]>(`http://localhost:8080/courses/${courseId}/tasks`, { headers }).subscribe({
      next: (data) => {
        this.tasks = data;
        console.log('✅ Aufgaben geladen:', data);
      },
      error: (err) => {
        console.error('❌ Fehler beim Laden der Tasks:', err);
      }
    });
  }

  openTask(task: any): void {
    localStorage.setItem('activeTask', JSON.stringify(task));
    window.location.href = '/codingSpace';
  }
  goBack(): void {
    window.location.href = '/courselist'
  }

}
