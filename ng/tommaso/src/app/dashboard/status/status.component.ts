import { Component, OnInit, ViewChild } from '@angular/core';
import { ReadDashboardService } from 'src/app/service/read-dashboard.service';
import { StoredProbeResults } from 'src/app/model/stored-probe-results';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

@Component({
  selector: 'app-status',
  templateUrl: './status.component.html',
  styleUrls: ['./status.component.scss']
})
export class StatusComponent implements OnInit {
  @ViewChild(MatSort, {static: true}) sort: MatSort;
  displayedColumns: string[] = ['source', 'target', 'status', 'elapsed', 'comment', 'time'];
  data: MatTableDataSource<StoredProbeResults>
  constructor(private svc: ReadDashboardService) { }

  ngOnInit(): void {
    this.svc.getAllResults().subscribe(rs => {
      this.data = new MatTableDataSource(rs);
      this.data.sort = this.sort;
    })

  }

  applyFilter(event: Event) {
    const filterValue = (event.target as HTMLInputElement).value;
    this.data.filter = filterValue.trim().toLowerCase();
  }

}
