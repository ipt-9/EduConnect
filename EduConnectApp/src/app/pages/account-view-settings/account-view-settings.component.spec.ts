import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AccountViewSettingsComponent } from './account-view-settings.component';

describe('AccountViewSettingsComponent', () => {
  let component: AccountViewSettingsComponent;
  let fixture: ComponentFixture<AccountViewSettingsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AccountViewSettingsComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AccountViewSettingsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
