import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CodingSpaceComponent } from './coding-space.component';

describe('CodingSpaceComponent', () => {
  let component: CodingSpaceComponent;
  let fixture: ComponentFixture<CodingSpaceComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CodingSpaceComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CodingSpaceComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
