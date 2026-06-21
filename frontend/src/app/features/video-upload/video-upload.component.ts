import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { VideosService } from '../../core/api';

@Component({
  selector: 'app-video-upload',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div style="max-width: 500px; margin: 40px auto; padding: 20px;">
      <h2>Upload de Vídeo</h2>

      <div style="margin-bottom: 16px;">
        <label>Arquivo:</label><br/>
        <input type="file" (change)="onFileSelected($event)" accept="video/*">
      </div>

      <div style="margin-bottom: 16px;">
        <label>Título:</label><br/>
        <input type="text" [(ngModel)]="titulo" placeholder="Título do vídeo"
               style="width: 100%; padding: 8px;">
      </div>

      <div style="margin-bottom: 16px;">
        <label>Autor:</label><br/>
        <input type="text" [(ngModel)]="autor" placeholder="Nome do autor"
               style="width: 100%; padding: 8px;">
      </div>

      <button (click)="upload()" [disabled]="!selectedFile || loading">
        {{ loading ? 'Enviando...' : 'Enviar' }}
      </button>

      <p *ngIf="message" style="margin-top: 16px;">{{ message }}</p>
    </div>
  `
})
export class VideoUploadComponent {
  selectedFile: File | null = null;
  titulo = '';
  autor = '';
  loading = false;
  message = '';

  constructor(
    private videosService: VideosService,
    private http: HttpClient
  ) {}

  onFileSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files?.length) {
      this.selectedFile = input.files[0];
    }
  }

  async upload() {
    if (!this.selectedFile || !this.titulo || !this.autor) return;

    this.loading = true;
    this.message = '';

    try {
      const uploadResponse = await firstValueFrom(
        this.videosService.createUploadUrl({
          fileName: this.selectedFile.name,
          fileType: this.selectedFile.type as any,
          fileSize: this.selectedFile.size
        })
      );

      await this.http.put(uploadResponse.uploadUrl, this.selectedFile, {
        headers: { 'Content-Type': this.selectedFile.type }
      }).toPromise();

      await firstValueFrom(
        this.videosService.createVideo({
          id: uploadResponse.videoId,
          titulo: this.titulo,
          autor: this.autor,
          url_s3: uploadResponse.uploadUrl.split('?')[0]
        })
      );

      this.message = 'Vídeo enviado com sucesso!';
      this.selectedFile = null;
      this.titulo = '';
      this.autor = '';
    } catch (error) {
      this.message = 'Erro ao enviar vídeo. Tente novamente.';
      console.error(error);
    } finally {
      this.loading = false;
    }
  }
}
