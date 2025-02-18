import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AccountViewLogsecurityComponent } from './account-view-logsecurity.component';

describe('AccountViewLogsecurityComponent', () => {
  let component: AccountViewLogsecurityComponent;
  let fixture: ComponentFixture<AccountViewLogsecurityComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AccountViewLogsecurityComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AccountViewLogsecurityComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
