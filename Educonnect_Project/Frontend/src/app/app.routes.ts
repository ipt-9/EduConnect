import { Routes } from '@angular/router';
import {LoginComponent} from './pages/login/login.component';
import {ContactComponent} from './pages/contact/contact.component';
import {ErrorpageComponent} from './pages/errorpage/errorpage.component';
import {HomepageComponent} from './pages/homepage/homepage.component';
import {RegisterComponent} from './pages/register/register.component';
import {CodingSpaceComponent} from './components/coding-space/coding-space.component';
import {CourseListComponent} from './components/course-list/course-list.component';
import {TasksListComponent} from './components/tasks-list/tasks-list.component';
import {AuthGuard} from './auth.guard';
import {SidebarComponent} from './components/sidebar/sidebar.component';
import {GroupManagerComponent} from './components/group-manager/group-manager.component';
import {GroupChatComponent} from './components/groupchat/groupchat.component';
import {GroupDetailsComponent} from './components/group-details/group-details.component';
import { GroupRoleManagerComponent } from './components/group-role-manager/group-role-manager.component';
import {DashboardComponent} from './components/dashboard/dashboard.component';
import {PaymentComponent} from './components/payment/payment.component';

export const routes: Routes = [
  { path: '', component: HomepageComponent },
  { path: 'codingSpace', component: CodingSpaceComponent, canActivate: [AuthGuard] },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'contact', component: ContactComponent, canActivate: [AuthGuard] },
  { path: 'errorpage', component: ErrorpageComponent, canActivate: [AuthGuard] },
  { path: 'courselist', component: CourseListComponent, canActivate: [AuthGuard] },
  { path: 'taskslist', component: TasksListComponent, canActivate: [AuthGuard] },
  { path: 'sidebar', component: SidebarComponent },
  { path: 'groupmanager',component: GroupManagerComponent, canActivate: [AuthGuard] },
  { path: 'groups/:id/chat',component: GroupChatComponent, canActivate: [AuthGuard] },
  { path: 'groups/:id', component: GroupDetailsComponent, canActivate: [AuthGuard] },
  { path: 'groups/:id/manage-roles', component: GroupRoleManagerComponent },
  { path: 'dashboard', component: DashboardComponent, canActivate: [AuthGuard] },
  {path: 'payment', component: PaymentComponent, canActivate: [AuthGuard]},
];
