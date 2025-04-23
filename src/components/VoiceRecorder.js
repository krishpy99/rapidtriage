import React, { useState, useEffect } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, ActivityIndicator } from 'react-native';
import AudioService from '../services/AudioService';
import { API_ENDPOINTS } from '../utils/config';

const VoiceRecorder = ({ onTranscriptReceived }) => {
  const [isRecording, setIsRecording] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [recordingDuration, setRecordingDuration] = useState(0);
  const [error, setError] = useState('');
  const [timer, setTimer] = useState(null);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timer) {
        clearInterval(timer);
      }
    };
  }, [timer]);

  const startRecording = async () => {
    try {
      setError('');
      const success = await AudioService.startRecording();
      
      if (success) {
        setIsRecording(true);
        setRecordingDuration(0);
        
        // Start timer
        const interval = setInterval(() => {
          setRecordingDuration(prev => prev + 1);
        }, 1000);
        
        setTimer(interval);
      } else {
        setError('Could not start recording');
      }
    } catch (err) {
      console.error('Error starting recording:', err);
      setError('Failed to start recording');
    }
  };

  const stopRecording = async () => {
    try {
      // Clear timer
      if (timer) {
        clearInterval(timer);
        setTimer(null);
      }
      
      setIsRecording(false);
      setIsProcessing(true);
      
      // Stop recording
      const recordingInfo = await AudioService.stopRecording();
      
      if (!recordingInfo) {
        setIsProcessing(false);
        setError('No recording to process');
        return;
      }
      
      // Send to backend
      const result = await AudioService.sendAudioToBackend(API_ENDPOINTS.CONVERT_AUDIO);
      
      // Process result
      if (result && result.success && result.transcript) {
        onTranscriptReceived(result.transcript);
      } else {
        setError(result?.error || 'Failed to transcribe audio');
      }
      
      setIsProcessing(false);
    } catch (err) {
      console.error('Error processing recording:', err);
      setError('Failed to process recording');
      setIsProcessing(false);
    }
  };

  // Format duration as MM:SS
  const formatDuration = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <View style={styles.container}>
      {!isRecording && !isProcessing ? (
        // Start recording button
        <TouchableOpacity 
          style={styles.recordButton} 
          onPress={startRecording}
        >
          <View style={styles.recordIcon} />
          <Text style={styles.buttonText}>Record Symptoms</Text>
        </TouchableOpacity>
      ) : isRecording ? (
        // Recording in progress
        <View style={styles.recordingContainer}>
          <Text style={styles.durationText}>{formatDuration(recordingDuration)}</Text>
          
          <View style={styles.pulseContainer}>
            <View style={[styles.pulse, styles.pulse1]} />
            <View style={[styles.pulse, styles.pulse2]} />
            <View style={[styles.pulse, styles.pulse3]} />
            <View style={styles.recordingDot} />
          </View>
          
          <TouchableOpacity 
            style={styles.stopButton} 
            onPress={stopRecording}
          >
            <View style={styles.stopIcon} />
            <Text style={styles.buttonText}>Stop Recording</Text>
          </TouchableOpacity>
        </View>
      ) : (
        // Processing recording
        <View style={styles.processingContainer}>
          <ActivityIndicator size="large" color="#0A2F52" />
          <Text style={styles.processingText}>Converting speech to text...</Text>
        </View>
      )}
      
      {error ? <Text style={styles.errorText}>{error}</Text> : null}
      
      <Text style={styles.instructionText}>
        Speak clearly and describe your symptoms in detail
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    padding: 16,
    backgroundColor: '#F0F0F0',
    borderRadius: 8,
    marginVertical: 12,
    alignItems: 'center',
  },
  recordButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#0A2F52',
    paddingVertical: 12,
    paddingHorizontal: 20,
    borderRadius: 25,
  },
  recordIcon: {
    width: 12,
    height: 12,
    borderRadius: 6,
    backgroundColor: '#FF4545',
    marginRight: 8,
  },
  buttonText: {
    color: 'white',
    fontWeight: '600',
    fontSize: 16,
  },
  recordingContainer: {
    alignItems: 'center',
    width: '100%',
  },
  durationText: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  pulseContainer: {
    width: 60,
    height: 60,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  pulse: {
    position: 'absolute',
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: '#FF4545',
    opacity: 0.2,
  },
  pulse1: {
    transform: [{ scale: 1 }],
    opacity: 0.3,
  },
  pulse2: {
    transform: [{ scale: 0.8 }],
    opacity: 0.4,
  },
  pulse3: {
    transform: [{ scale: 0.6 }],
    opacity: 0.5,
  },
  recordingDot: {
    width: 20,
    height: 20,
    borderRadius: 10,
    backgroundColor: '#FF4545',
  },
  stopButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#D32F2F',
    paddingVertical: 12,
    paddingHorizontal: 20,
    borderRadius: 25,
  },
  stopIcon: {
    width: 12,
    height: 12,
    backgroundColor: 'white',
    marginRight: 8,
  },
  processingContainer: {
    alignItems: 'center',
    padding: 20,
  },
  processingText: {
    marginTop: 10,
    fontSize: 16,
    color: '#555',
  },
  errorText: {
    color: '#D32F2F',
    marginTop: 12,
    textAlign: 'center',
  },
  instructionText: {
    marginTop: 16,
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
  },
});

export default VoiceRecorder;
