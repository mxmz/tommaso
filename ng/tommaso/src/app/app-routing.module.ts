import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { StatusComponent } from './dashboard/status/status.component';


const routes: Routes = [{
   path: '', component: StatusComponent, pathMatch: 'full' },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
