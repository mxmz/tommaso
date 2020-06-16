import { TestBed } from '@angular/core/testing';

import { ReadDashboardService } from './read-dashboard.service';

describe('ReadDashboardService', () => {
  let service: ReadDashboardService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ReadDashboardService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
