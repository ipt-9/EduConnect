import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AccountViewNotificationsComponent } from './account-view-notifications.component';

describe('AccountViewNotificationsComponent', () => {
  let component: AccountViewNotificationsComponent;
  let fixture: ComponentFixture<AccountViewNotificationsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AccountViewNotificationsComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AccountViewNotificationsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
