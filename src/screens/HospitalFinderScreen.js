import React, { useState } from 'react';
import { 
  View, 
  Text, 
  StyleSheet, 
  SafeAreaView, 
  ScrollView,
  TouchableOpacity
} from 'react-native';
import NearbyHospitals from '../components/hospitals/NearbyHospitals';
import HospitalMap from '../components/hospitals/HospitalMap';

const HospitalFinderScreen = ({ navigation }) => {
  const [userLocation, setUserLocation] = useState(null);
  const [hospitals, setHospitals] = useState([]);
  const [selectedHospital, setSelectedHospital] = useState(null);

  const handleHospitalSelect = (hospital) => {
    setSelectedHospital(hospital);
  };

  const handleUseSelectedHospital = () => {
    if (selectedHospital) {
      navigation.navigate('EmergencyResults', {
        hospital: selectedHospital
      });
    }
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
          <Text style={styles.subtitle}>Find Nearby Hospitals</Text>
          
          {userLocation && hospitals.length > 0 && (
            <HospitalMap 
              userLocation={userLocation}
              hospitals={hospitals}
              selectedHospital={selectedHospital}
            />
          )}
          
          <NearbyHospitals 
            onSelectHospital={handleHospitalSelect}
            onLocationReceived={(location) => setUserLocation(location)}
            onHospitalsReceived={(data) => setHospitals(data)}
          />
          
          {selectedHospital && (
            <TouchableOpacity 
              style={styles.useHospitalButton}
              onPress={handleUseSelectedHospital}
            >
              <Text style={styles.useHospitalButtonText}>
                Use {selectedHospital.name}
              </Text>
            </TouchableOpacity>
          )}
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
  useHospitalButton: {
    backgroundColor: '#0A2F52',
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    marginTop: 20,
  },
  useHospitalButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: 'bold',
  },
});

export default HospitalFinderScreen;
