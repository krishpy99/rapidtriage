import * as FileSystem from 'expo-file-system';
import { Audio } from 'expo-av';
import { Platform } from 'react-native';

class AudioService {
  constructor() {
    this.recording = null;
    this.sound = null;
    this.uri = null;
  }

  // Initialize audio recording
  async init() {
    try {
      await Audio.requestPermissionsAsync();
      
      await Audio.setAudioModeAsync({
        allowsRecordingIOS: true,
        playsInSilentModeIOS: true,
        staysActiveInBackground: false,
        shouldDuckAndroid: true,
        playThroughEarpieceAndroid: false,
      });
      
      return true;
    } catch (error) {
      console.error('Failed to initialize audio:', error);
      return false;
    }
  }

  // Start recording
  async startRecording() {
    try {
      const initSuccess = await this.init();
      if (!initSuccess) return false;
      
      const { recording } = await Audio.Recording.createAsync(
        Audio.RecordingOptionsPresets.HIGH_QUALITY
      );
      
      this.recording = recording;
      return true;
    } catch (error) {
      console.error('Failed to start recording:', error);
      return false;
    }
  }

  // Stop recording
  async stopRecording() {
    try {
      if (!this.recording) return null;
      
      await this.recording.stopAndUnloadAsync();
      
      const uri = this.recording.getURI();
      this.uri = uri;
      
      const fileInfo = await FileSystem.getInfoAsync(uri);
      
      this.recording = null;
      
      return {
        uri,
        size: fileInfo.size,
        duration: fileInfo.modificationTime || 0,
      };
    } catch (error) {
      console.error('Failed to stop recording:', error);
      return null;
    }
  }

  // Send audio to backend for processing
  async sendAudioToBackend(apiUrl) {
    try {
      if (!this.uri) throw new Error('No recording available');
      
      const formData = new FormData();
      
      // Determine file type based on platform
      const fileType = Platform.OS === 'ios' ? 'audio/m4a' : 'audio/aac';
      const fileName = Platform.OS === 'ios' ? 'recording.m4a' : 'recording.aac';
      
      formData.append('audio', {
        uri: this.uri,
        name: fileName,
        type: fileType,
      });
      
      const response = await fetch(apiUrl, {
        method: 'POST',
        body: formData,
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      
      if (!response.ok) {
        throw new Error(`Server responded with ${response.status}`);
      }
      
      const result = await response.json();
      return result;
    } catch (error) {
      console.error('Failed to send audio to backend:', error);
      throw error;
    }
  }
}

export default new AudioService();
