import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

// Web version of HospitalMap that doesn't use react-native-maps
const HospitalMapWeb = ({ userLocation, hospitals, selectedHospital }) => {
  return (
    <View style={styles.container}>
      <Text style={styles.mapText}>Map View</Text>
      <Text style={styles.mapSubtext}>
        Map display is available on mobile devices.
        {userLocation ? ` Your location: ${userLocation.latitude.toFixed(4)}, ${userLocation.longitude.toFixed(4)}` : ''}
      </Text>
      {hospitals && hospitals.length > 0 && (
        <View style={styles.hospitalsContainer}>
          <Text style={styles.hospitalsTitle}>Nearby Hospitals:</Text>
          {hospitals.map((hospital, index) => (
            <Text key={hospital.place_id || index} style={[
              styles.hospitalItem,
              selectedHospital && selectedHospital.place_id === hospital.place_id ? styles.selectedHospital : {}
            ]}>
              {hospital.name} - {hospital.vicinity}
            </Text>
          ))}
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    height: 200,
    width: '100%',
    borderRadius: 8,
    backgroundColor: '#E5E5E5',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 16,
    marginVertical: 16,
  },
  mapText: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  mapSubtext: {
    fontSize: 14,
    textAlign: 'center',
    color: '#555',
  },
  hospitalsContainer: {
    marginTop: 12,
    width: '100%',
  },
  hospitalsTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  hospitalItem: {
    fontSize: 14,
    marginVertical: 2,
    padding: 4,
  },
  selectedHospital: {
    backgroundColor: '#EAF2F8',
    fontWeight: 'bold',
    borderRadius: 4,
  },
});

export default HospitalMapWeb;
