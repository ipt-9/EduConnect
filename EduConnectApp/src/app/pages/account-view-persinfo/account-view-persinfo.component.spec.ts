import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AccountViewPersinfoComponent } from './account-view-persinfo.component';

describe('AccountViewPersinfoComponent', () => {
  let component: AccountViewPersinfoComponent;
  let fixture: ComponentFixture<AccountViewPersinfoComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AccountViewPersinfoComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AccountViewPersinfoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
