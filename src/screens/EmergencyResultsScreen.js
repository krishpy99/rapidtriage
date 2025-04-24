import React from 'react';
import { 
  View, 
  Text, 
  TouchableOpacity, 
  StyleSheet, 
  SafeAreaView, 
  Linking,
  Platform 
} from 'react-native';

const EmergencyResultsScreen = ({ navigation, route }) => {
  // Get hospital data from route params if available
  const hospital = route.params?.hospital || {
    name: 'General Hospital',
    address: '123 Health St.',
  };

  const handleCall911 = () => {
    const phoneNumber = Platform.OS === 'ios' ? 'telprompt:911' : 'tel:911';
    Linking.openURL(phoneNumber).catch(err => {
      console.error('Failed to open dialer:', err);
    });
  };

  const handleCallHospital = () => {
    if (hospital.phone) {
      const phoneNumber = Platform.OS === 'ios' ? `telprompt:${hospital.phone}` : `tel:${hospital.phone}`;
      Linking.openURL(phoneNumber).catch(err => {
        console.error('Failed to open dialer:', err);
      });
    } else {
      // If no phone number, call 911
      handleCall911();
    }
  };

  const handleGetDirections = () => {
    if (hospital.location) {
      const { latitude, longitude } = hospital.location;
      const label = encodeURIComponent(hospital.name);
      
      // Open in Google Maps or Apple Maps
      const url = Platform.select({
        ios: `maps:0,0?q=${label}@${latitude},${longitude}`,
        android: `geo:0,0?q=${latitude},${longitude}(${label})`,
      });
      
      Linking.openURL(url);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      
      <View style={styles.content}>
        <Text style={styles.title}>RapidTriage AI</Text>
        
        <Text style={styles.resultTitle}>Emergency results</Text>
        
        <Text style={styles.resultMessage}>
          Based on the analysis, your symptoms indicate{'\n'}
          that you may need emergency care.
        </Text>
        
        <View style={styles.hospitalCard}>
          <View style={styles.hospitalIcon}>
            <Text style={styles.hospitalIconPlus}>+</Text>
            <View style={styles.hospitalIconGrid}>
              {[...Array(9)].map((_, i) => (
                <View key={i} style={styles.hospitalIconSquare} />
              ))}
            </View>
          </View>
          <View style={styles.hospitalInfo}>
            <Text style={styles.hospitalName}>{hospital.name}</Text>
            <Text style={styles.hospitalAddress}>{hospital.address}</Text>
          </View>
        </View>
        
        <TouchableOpacity 
          style={styles.callButton}
          onPress={handleCall911}
        >
          <Text style={styles.callButtonText}>Call 911</Text>
        </TouchableOpacity>
        
        {hospital.location && (
          <View style={styles.additionalActions}>
            <TouchableOpacity 
              style={styles.secondaryButton}
              onPress={handleCallHospital}
            >
              <Text style={styles.secondaryButtonText}>Call Hospital</Text>
            </TouchableOpacity>
            
            <TouchableOpacity 
              style={styles.secondaryButton}
              onPress={handleGetDirections}
            >
              <Text style={styles.secondaryButtonText}>Get Directions</Text>
            </TouchableOpacity>
          </View>
        )}
        
        <TouchableOpacity 
          style={styles.backLink}
          onPress={() => navigation.navigate('Assessment')}
        >
          <Text style={styles.backLinkText}>Go back to symptoms</Text>
        </TouchableOpacity>
        
        <Text style={styles.disclaimer}>
          For informational purposes only.{'\n'}
          Not a substitute for professional medical advice.
        </Text>
      </View>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  content: {
    flex: 1,
    padding: 20,
    alignItems: 'center',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333333',
    marginBottom: 20,
    textAlign: 'center',
  },
  resultTitle: {
    fontSize: 26,
    fontWeight: 'bold',
    color: '#333333',
    marginBottom: 16,
  },
  resultMessage: {
    fontSize: 18,
    color: '#555555',
    textAlign: 'center',
    marginBottom: 30,
    lineHeight: 24,
  },
  hospitalCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#DDDDDD',
    padding: 16,
    width: '100%',
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 30,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  hospitalIcon: {
    backgroundColor: '#0A2F52',
    borderRadius: 8,
    width: 60,
    height: 60,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  hospitalIconPlus: {
    color: '#FFFFFF',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  hospitalIconGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    width: 36,
    height: 18,
  },
  hospitalIconSquare: {
    width: 6,
    height: 6,
    backgroundColor: '#FFFFFF',
    margin: 1,
  },
  hospitalInfo: {
    flex: 1,
  },
  hospitalName: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#333333',
    marginBottom: 4,
  },
  hospitalAddress: {
    fontSize: 16,
    color: '#666666',
  },
  callButton: {
    backgroundColor: '#0A2F52',
    borderRadius: 25,
    paddingVertical: 14,
    paddingHorizontal: 20,
    width: '100%',
    alignItems: 'center',
    marginBottom: 16,
  },
  callButtonText: {
    color: '#FFFFFF',
    fontSize: 18,
    fontWeight: 'bold',
  },
  additionalActions: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    width: '100%',
    marginBottom: 16,
  },
  secondaryButton: {
    backgroundColor: '#EAF2F8',
    borderWidth: 1,
    borderColor: '#0A2F52',
    borderRadius: 25,
    paddingVertical: 10,
    paddingHorizontal: 15,
    flex: 0.48,
    alignItems: 'center',
  },
  secondaryButtonText: {
    color: '#0A2F52',
    fontSize: 16,
    fontWeight: '600',
  },
  backLink: {
    marginVertical: 16,
  },
  backLinkText: {
    color: '#0A2F52',
    fontSize: 16,
    textDecorationLine: 'underline',
  },
  disclaimer: {
    position: 'absolute',
    bottom: 20,
    left: 0,
    right: 0,
    textAlign: 'center',
    color: '#777777',
    fontSize: 12,
    lineHeight: 18,
  },
});

export default EmergencyResultsScreen;
