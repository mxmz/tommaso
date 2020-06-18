import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { StoredProbeResults } from '../model/stored-probe-results';

@Injectable({
  providedIn: 'root'
})
export class ReadDashboardService {

  constructor(private http: HttpClient,) { }


  getAllResults(filter: string): Observable<StoredProbeResults[]> {
    return this.http.get<StoredProbeResults[]>('/api/dashboard/probe/results', { params: { filter: filter } });
  }
}
