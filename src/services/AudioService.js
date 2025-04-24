import * as FileSystem from 'expo-file-system';
import { Audio } from 'expo-av';
import { Platform } from 'react-native';
import LocationService from './LocationService';

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
      
      // Use WAV format for both platforms
      const { recording } = await Audio.Recording.createAsync({
        android: {
          extension: '.wav',
          outputFormat: Audio.RECORDING_OPTION_ANDROID_OUTPUT_FORMAT_DEFAULT,
          audioEncoder: Audio.RECORDING_OPTION_ANDROID_AUDIO_ENCODER_DEFAULT,
          sampleRate: 44100,
          numberOfChannels: 2,
          bitRate: 128000,
        },
        ios: {
          extension: '.wav',
          outputFormat: Audio.RECORDING_OPTION_IOS_OUTPUT_FORMAT_LINEARPCM,
          audioQuality: Audio.RECORDING_OPTION_IOS_AUDIO_QUALITY_HIGH,
          sampleRate: 44100,
          numberOfChannels: 1,
          bitRate: 128000,
          linearPCMBitDepth: 16,
          linearPCMIsBigEndian: false,
          linearPCMIsFloat: false,
        },
      });
      
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
    console.log('Sending audio to backend:', apiUrl);
    try {
      if (!this.uri) throw new Error('No recording available');
      
      const formData = new FormData();
      
      // Use WAV format for both platforms
      const fileType = 'audio/wav';
      const fileName = 'recording.wav';
      
      formData.append('audio', {
        uri: this.uri,
        name: fileName,
        type: fileType,
      });

      // Get current location and add to request
      try {
        const location = await LocationService.getCurrentLocation();
        
        // Format location as required by the API
        const locationData = {
          latitude: location.latitude,
          longitude: location.longitude,
          address: 'Current location' // This could be improved with reverse geocoding
        };
        
        formData.append('location', JSON.stringify(locationData));
      } catch (locError) {
        console.warn('Could not get location for emergency request:', locError);
        // Continue without location if we can't get it
      }

      console.log('DEBUG Form data prepared:', formData);
      
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
      
      // Format the emergency response for display
      const formattedResponse = this.formatEmergencyResponse(result);
      return formattedResponse;
    } catch (error) {
      console.error('Failed to send audio to backend:', error);
      throw error;
    }
  }

  /**
   * Format the emergency response for display
   * @param {Object} response - Raw emergency response data
   * @returns {Object} - Formatted response with message property
   */
  formatEmergencyResponse(response) {
    if (!response) return { message: "No response received from emergency services." };

    // Extract the main components
    const { emergency_id, code, summary, timestamp, tool_responses } = response;
    
    // Build a formatted message from the summary
    let formattedMessage = summary || "Emergency response received.";
    
    // Add information about tools that were called
    if (tool_responses && tool_responses.length > 0) {
      formattedMessage += '\n\n🚨 Actions taken:';
      
      tool_responses.forEach(tool => {
        const status = tool.success ? '✅' : '❌';
        formattedMessage += `\n${status} ${tool.tool_name}`;
      });
    }
    
    // Add emergency ID for reference
    if (emergency_id) {
      formattedMessage += `\n\nEmergency ID: ${emergency_id}`;
    }
    
    // Return the original response with an added message property for compatibility
    return {
      ...response,
      message: formattedMessage
    };
  }
}

export default new AudioService();
