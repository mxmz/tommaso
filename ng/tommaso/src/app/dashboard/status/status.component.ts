import { Component, OnInit, ViewChild } from '@angular/core';
import { ReadDashboardService } from 'src/app/service/read-dashboard.service';
import { StoredProbeResults } from 'src/app/model/stored-probe-results';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { ActivatedRoute, Router, } from '@angular/router';
import { Location } from '@angular/common';
import { timer } from 'rxjs';
import { not } from '@angular/compiler/src/output/output_ast';


@Component({
  selector: 'app-status',
  templateUrl: './status.component.html',
  styleUrls: ['./status.component.scss']
})
export class StatusComponent implements OnInit {
  @ViewChild(MatSort, { static: true }) sort: MatSort;
  displayedColumns: string[] = ['source', 'target', 'description', 'passed', 'status', 'elapsed', 'comment', 'time'];
  data = new MatTableDataSource<StoredProbeResults>([]);
  loading = 0;
  filter = "";
  updated = new Date();
  constructor(private svc: ReadDashboardService, private route: ActivatedRoute, private location: Location, private r: Router) { }

  loadData() {
    this.loading++;
    const filter = this.filter;
    this.svc.getAllResults(filter).subscribe(rs => {
      this.data = new MatTableDataSource<StoredProbeResults>(rs);
      this.data.sort = this.sort;
      this.data.filterPredicate = (data, filter) => data.source.includes(filter) || data.args[0].includes(filter);
      this.applyFilter();
      this.loading--;
      //this.location.go(`?filter=${filter}`);
      this.updated = new Date();
    })
  }

  ngOnInit(): void {

    this.route.queryParamMap.subscribe(
      pm => {
        this.filter = pm.get('filter') || "";
        //this.applyFilter();
        this.loadData();

      });

    const source = timer(1000, 2000);
    let filter = this.filter;

    const abc = source.subscribe(val => {
      //console.log(val, '-');
      const now = new Date();
      if (!this.loading) {
        if (this.filter !== filter) {
          filter = this.filter;
          this.r.navigate([], { queryParams: { filter: filter, } })
        } else if (now.getTime() - this.updated.getTime() > 60000) {
          this.loadData();
        }
      }
    });

  }

  setFilter(event: Event) {
    const filter = (event.target as HTMLInputElement).value.trim().toLowerCase();
    if (filter !== this.filter) {
      this.filter = filter;
      this.applyFilter();
    }
  }

  applyFilter() {
    this.data.filter = this.filter

  }

}
