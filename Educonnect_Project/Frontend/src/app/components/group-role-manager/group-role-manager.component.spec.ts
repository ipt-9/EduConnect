import { ComponentFixture, TestBed } from '@angular/core/testing';

import { GroupRoleManagerComponent } from './group-role-manager.component';

describe('GroupRoleManagerComponent', () => {
  let component: GroupRoleManagerComponent;
  let fixture: ComponentFixture<GroupRoleManagerComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GroupRoleManagerComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(GroupRoleManagerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
