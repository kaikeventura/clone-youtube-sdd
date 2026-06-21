import { Component } from '@angular/core';
import { VideoUploadComponent } from './features/video-upload/video-upload.component';
import { VideoListComponent } from './features/video-list/video-list.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [VideoUploadComponent, VideoListComponent],
  template: `
    <app-video-upload />
    <hr>
    <app-video-list />
  `,
  styles: [],
})
export class AppComponent {
  title = 'frontend';
}
