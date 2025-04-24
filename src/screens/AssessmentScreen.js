import React, { useState } from 'react';
import { 
  View, 
  Text, 
  TextInput, 
  TouchableOpacity, 
  StyleSheet, 
  ScrollView, 
  SafeAreaView 
} from 'react-native';
import { Picker } from '@react-native-picker/picker';
import Slider from '@react-native-community/slider';
import VoiceRecorder from '../components/VoiceRecorder';
import { API_ENDPOINTS } from '../utils/config';

const AssessmentScreen = ({ navigation }) => {
  const [symptoms, setSymptoms] = useState('');
  const [duration, setDuration] = useState('less-than-24');
  const [painLevel, setPainLevel] = useState(5);
  const [showVoiceRecorder, setShowVoiceRecorder] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async () => {
    try {
      setIsSubmitting(true);
      setError('');
      
      // For now, just navigate to the emergency screen
      setTimeout(() => {
        setIsSubmitting(false);
        navigation.navigate('EmergencyResults');
      }, 1500);
      
    } catch (err) {
      console.error('Error submitting assessment:', err);
      setError('Failed to submit assessment. Please try again.');
      setIsSubmitting(false);
    }
  };

  const handleTranscript = (transcript) => {
    // Append transcript to symptoms text
    if (symptoms.trim() === '') {
      setSymptoms(transcript);
    } else {
      setSymptoms(prevSymptoms => `${prevSymptoms}\n\n${transcript}`);
    }
    
    // Hide voice recorder
    setShowVoiceRecorder(false);
  };

  const handleFindHospitals = () => {
    navigation.navigate('HospitalFinder');
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView style={styles.scrollView}>
        <View style={styles.content}>
          <TouchableOpacity 
            style={styles.backButton}
            onPress={() => navigation.goBack()}
          >
            <Text style={styles.backButtonText}>← Back</Text>
          </TouchableOpacity>
          
          <Text style={styles.title}>RapidTriage AI</Text>
          <Text style={styles.subtitle}>Symptom Assessment</Text>
          
          <View style={styles.formGroup}>
            <Text style={styles.label}>What symptoms are you experiencing?</Text>
            <View style={styles.inputHeader}>
              <Text style={styles.inputHint}>Type or use voice input</Text>
              <TouchableOpacity onPress={() => setShowVoiceRecorder(!showVoiceRecorder)}>
                <Text style={styles.voiceButton}>
                  {showVoiceRecorder ? 'Hide voice input' : 'Use voice input'}
                </Text>
              </TouchableOpacity>
            </View>
            <TextInput
              style={styles.textArea}
              multiline
              numberOfLines={4}
              placeholder="Describe your symptoms here..."
              value={symptoms}
              onChangeText={setSymptoms}
            />
            
            {showVoiceRecorder && (
              <VoiceRecorder onTranscriptReceived={handleTranscript} />
            )}
          </View>
          
          <View style={styles.formGroup}>
            <Text style={styles.label}>How long have you had these symptoms?</Text>
            <View style={styles.pickerContainer}>
              <Picker
                selectedValue={duration}
                style={styles.picker}
                onValueChange={(itemValue) => setDuration(itemValue)}
              >
                <Picker.Item label="Less than 24 hours" value="less-than-24" />
                <Picker.Item label="1-3 days" value="1-3-days" />
                <Picker.Item label="3-7 days" value="3-7-days" />
                <Picker.Item label="More than a week" value="more-than-week" />
              </Picker>
            </View>
          </View>
          
          <View style={styles.formGroup}>
            <Text style={styles.label}>Rate your pain level (0-10)</Text>
            <Slider
              style={styles.slider}
              minimumValue={0}
              maximumValue={10}
              step={1}
              value={painLevel}
              onValueChange={setPainLevel}
              minimumTrackTintColor="#0A2F52"
              maximumTrackTintColor="#DDDDDD"
              thumbTintColor="#0A2F52"
            />
            <View style={styles.sliderLabels}>
              <Text style={styles.sliderLabel}>No pain</Text>
              <Text style={styles.sliderLabel}>Severe pain</Text>
            </View>
          </View>
          
          <TouchableOpacity 
            style={styles.hospitalFinderButton}
            onPress={handleFindHospitals}
          >
            <Text style={styles.hospitalFinderButtonText}>Find Nearby Hospitals</Text>
          </TouchableOpacity>
          
          {error ? <Text style={styles.errorText}>{error}</Text> : null}
          
          <TouchableOpacity 
            style={[styles.button, isSubmitting && styles.buttonDisabled]}
            onPress={handleSubmit}
            disabled={isSubmitting}
          >
            <Text style={styles.buttonText}>
              {isSubmitting ? 'Analyzing...' : 'Analyze Symptoms'}
            </Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  scrollView: {
    flex: 1,
  },
  content: {
    padding: 20,
    paddingBottom: 40,
  },
  backButton: {
    marginBottom: 20,
  },
  backButtonText: {
    color: '#0A2F52',
    fontSize: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333333',
    marginBottom: 10,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#555555',
    marginBottom: 20,
    textAlign: 'center',
  },
  formGroup: {
    marginBottom: 24,
  },
  label: {
    fontSize: 16,
    color: '#444444',
    marginBottom: 8,
  },
  inputHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  inputHint: {
    fontSize: 12,
    color: '#777777',
  },
  voiceButton: {
    fontSize: 14,
    color: '#0A2F52',
    fontWeight: '500',
  },
  textArea: {
    backgroundColor: '#FFFFFF',
    borderWidth: 1,
    borderColor: '#DDDDDD',
    borderRadius: 8,
    padding: 12,
    fontSize: 16,
    minHeight: 100,
    textAlignVertical: 'top',
  },
  pickerContainer: {
    borderWidth: 1,
    borderColor: '#DDDDDD',
    borderRadius: 8,
    backgroundColor: '#FFFFFF',
    overflow: 'hidden',
  },
  picker: {
    height: 50,
  },
  slider: {
    width: '100%',
    height: 40,
  },
  sliderLabels: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  sliderLabel: {
    color: '#777777',
    fontSize: 12,
  },
  hospitalFinderButton: {
    backgroundColor: '#EAF2F8',
    borderWidth: 1,
    borderColor: '#0A2F52',
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
    marginBottom: 20,
  },
  hospitalFinderButtonText: {
    color: '#0A2F52',
    fontSize: 16,
    fontWeight: '600',
  },
  errorText: {
    color: '#D32F2F',
    marginBottom: 12,
  },
  button: {
    backgroundColor: '#0A2F52',
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
    marginTop: 10,
  },
  buttonDisabled: {
    backgroundColor: '#0A2F52',
    opacity: 0.7,
  },
  buttonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: 'bold',
  },
});

export default AssessmentScreen;
