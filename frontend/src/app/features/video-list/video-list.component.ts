import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { VideosService, Video } from '../../core/api';

@Component({
  selector: 'app-video-list',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div style="max-width: 800px; margin: 20px auto; padding: 20px;">
      <h2>Vídeos</h2>
      <div *ngIf="videos.length === 0">Nenhum vídeo encontrado.</div>
      <div *ngFor="let video of videos" style="border: 1px solid #ccc; border-radius: 8px; padding: 16px; margin-bottom: 16px;">
        <h3>{{ video.titulo }}</h3>
        <p><strong>Autor:</strong> {{ video.autor }}</p>
        <video controls width="100%" [src]="video.url_s3"></video>
      </div>
    </div>
  `
})
export class VideoListComponent implements OnInit {
  videos: Video[] = [];

  constructor(private videosService: VideosService) {}

  ngOnInit() {
    this.videosService.listVideos().subscribe({
      next: (res) => this.videos = res.videos,
      error: (err) => console.error('Erro ao listar vídeos:', err)
    });
  }
}
