import { Component, OnInit } from '@angular/core';
import { HttpClient, HttpHeaders, HttpClientModule } from '@angular/common/http';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { RouterModule } from '@angular/router';


@Component({
  selector: 'app-course-list',
  standalone: true,
  imports: [HttpClientModule, CommonModule,SidebarComponent,RouterModule],
  templateUrl: './course-list.component.html',
  styleUrls: ['./course-list.component.scss']
})
export class CourseListComponent implements OnInit {
  courses: any[] = [];

  constructor(private http: HttpClient, private router: Router) {}

  ngOnInit(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<any[]>('http://localhost:8080/my-courses', { headers }).subscribe({
      next: (data) => {
        this.courses = data;
        console.log('✅ Kurse geladen:', data);
      },
      error: (err) => {
        console.error('❌ Fehler beim Laden der Kurse:', err);
      }
    });
  }

  openCourse(courseId: number): void {
    // 🧠 Kurs-ID im localStorage merken (optional)
    localStorage.setItem('activeCourseId', courseId.toString());

    // 🚀 Weiterleitung zur Aufgaben-Ansicht
    this.router.navigate(['/taskslist']);
  }
}
