import React, { useState, useEffect } from 'react';
import { 
  View, 
  Text, 
  StyleSheet, 
  TouchableOpacity, 
  FlatList,
  ActivityIndicator,
  Linking,
  Platform
} from 'react-native';

const NearbyHospitals = ({ onSelectHospital, onLocationReceived, onHospitalsReceived }) => {
  const [loading, setLoading] = useState(true);
  const [hospitals, setHospitals] = useState([]);
  const [error, setError] = useState(null);
  const [userLocation, setUserLocation] = useState(null);

  useEffect(() => {
    // Simulate fetching location and hospitals for demo
    setTimeout(() => {
      const mockLocation = {
        latitude: 37.7749,
        longitude: -122.4194
      };
      setUserLocation(mockLocation);
      
      if (onLocationReceived) {
        onLocationReceived(mockLocation);
      }
      
      const mockHospitals = [
        {
          place_id: 'hospital1',
          name: 'General Hospital',
          vicinity: '123 Health St, San Francisco',
          geometry: {
            location: {
              lat: 37.7739,
              lng: -122.4312
            }
          },
          rating: 4.2,
          opening_hours: { open_now: true }
        },
        {
          place_id: 'hospital2',
          name: 'City Medical Center',
          vicinity: '456 Medical Ave, San Francisco',
          geometry: {
            location: {
              lat: 37.7850,
              lng: -122.4260
            }
          },
          rating: 4.5,
          opening_hours: { open_now: false }
        },
        {
          place_id: 'hospital3',
          name: 'Community Hospital',
          vicinity: '789 Care Blvd, San Francisco',
          geometry: {
            location: {
              lat: 37.7695,
              lng: -122.4100
            }
          },
          rating: 3.9,
          opening_hours: { open_now: true }
        }
      ];
      
      setHospitals(mockHospitals);
      
      if (onHospitalsReceived) {
        onHospitalsReceived(mockHospitals);
      }
      
      setLoading(false);
    }, 1500);
  }, []);

  const formatDistance = (hospital) => {
    if (!userLocation) return '';
    
    const distance = calculateDistance(
      userLocation.latitude,
      userLocation.longitude,
      hospital.geometry.location.lat,
      hospital.geometry.location.lng
    );
    
    return distance < 1 
      ? `${(distance * 1000).toFixed(0)} m` 
      : `${distance.toFixed(1)} km`;
  };

  const calculateDistance = (lat1, lon1, lat2, lon2) => {
    const R = 6371; // Radius of the earth in km
    const dLat = deg2rad(lat2 - lat1);
    const dLon = deg2rad(lon2 - lon1);
    const a = 
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(deg2rad(lat1)) * Math.cos(deg2rad(lat2)) * 
      Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    const distance = R * c; // Distance in km
    return distance;
  };
  
  const deg2rad = (deg) => {
    return deg * (Math.PI / 180);
  };

  const handleSelectHospital = (hospital) => {
    if (onSelectHospital) {
      onSelectHospital({
        place_id: hospital.place_id,
        name: hospital.name,
        address: hospital.vicinity,
        location: {
          latitude: hospital.geometry.location.lat,
          longitude: hospital.geometry.location.lng,
        },
        rating: hospital.rating,
        open_now: hospital.opening_hours ? hospital.opening_hours.open_now : null,
      });
    }
  };

  const handleOpenMaps = (hospital) => {
    // For demo purposes, just show an alert on web
    if (Platform.OS === 'web') {
      alert(`Would open directions to: ${hospital.name}`);
      return;
    }
    
    const { lat, lng } = hospital.geometry.location;
    const label = encodeURIComponent(hospital.name);
    
    // Open in Google Maps or Apple Maps
    const url = Platform.select({
      ios: `maps:0,0?q=${label}@${lat},${lng}`,
      android: `geo:0,0?q=${lat},${lng}(${label})`,
    });
    
    Linking.openURL(url);
  };

  const handleCallHospital = (hospital) => {
    // For demo purposes, just show an alert on web
    if (Platform.OS === 'web') {
      alert(`Would call: ${hospital.name}`);
      return;
    }
    
    // Check if phone number is available
    if (hospital.formatted_phone_number) {
      Linking.openURL(`tel:${hospital.formatted_phone_number}`);
    } else {
      alert('Phone number not available. Please check their website.');
    }
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#0A2F52" />
        <Text style={styles.loadingText}>Finding hospitals near you...</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.errorContainer}>
        <Text style={styles.errorText}>{error}</Text>
        <TouchableOpacity 
          style={styles.retryButton}
          onPress={() => setLoading(true)}
        >
          <Text style={styles.retryButtonText}>Retry</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Nearby Hospitals</Text>
      
      {hospitals.length === 0 ? (
        <Text style={styles.noResultsText}>No hospitals found nearby.</Text>
      ) : (
        <FlatList
          data={hospitals}
          keyExtractor={(item) => item.place_id}
          renderItem={({ item }) => (
            <TouchableOpacity 
              style={styles.hospitalItem}
              onPress={() => handleSelectHospital(item)}
            >
              <View style={styles.hospitalInfo}>
                <Text style={styles.hospitalName}>{item.name}</Text>
                <Text style={styles.hospitalAddress}>{item.vicinity}</Text>
                <View style={styles.hospitalDetails}>
                  <Text style={styles.hospitalDistance}>{formatDistance(item)}</Text>
                  {item.rating && (
                    <Text style={styles.hospitalRating}>⭐ {item.rating.toFixed(1)}</Text>
                  )}
                  {item.opening_hours && (
                    <Text style={[
                      styles.hospitalOpenStatus,
                      { color: item.opening_hours.open_now ? '#4CAF50' : '#F44336' }
                    ]}>
                      {item.opening_hours.open_now ? 'Open' : 'Closed'}
                    </Text>
                  )}
                </View>
              </View>
              
              <View style={styles.hospitalActions}>
                <TouchableOpacity 
                  style={styles.actionButton}
                  onPress={() => handleOpenMaps(item)}
                >
                  <Text style={styles.actionButtonText}>Directions</Text>
                </TouchableOpacity>
                
                <TouchableOpacity 
                  style={styles.actionButton}
                  onPress={() => handleCallHospital(item)}
                >
                  <Text style={styles.actionButtonText}>Call</Text>
                </TouchableOpacity>
              </View>
            </TouchableOpacity>
          )}
        />
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
  },
  title: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 16,
    color: '#0A2F52',
  },
  loadingContainer: {
    padding: 20,
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 10,
    fontSize: 16,
    color: '#555',
  },
  errorContainer: {
    padding: 20,
    alignItems: 'center',
  },
  errorText: {
    color: '#D32F2F',
    marginBottom: 16,
    textAlign: 'center',
  },
  retryButton: {
    backgroundColor: '#0A2F52',
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderRadius: 4,
  },
  retryButtonText: {
    color: 'white',
    fontWeight: '600',
  },
  noResultsText: {
    textAlign: 'center',
    color: '#666',
    marginTop: 20,
  },
  hospitalItem: {
    backgroundColor: 'white',
    borderRadius: 8,
    marginBottom: 12,
    padding: 12,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
  },
  hospitalInfo: {
    flex: 1,
    marginBottom: 8,
  },
  hospitalName: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 4,
  },
  hospitalAddress: {
    fontSize: 14,
    color: '#666',
    marginBottom: 8,
  },
  hospitalDetails: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  hospitalDistance: {
    fontSize: 12,
    color: '#666',
    marginRight: 12,
    backgroundColor: '#F0F0F0',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
  },
  hospitalRating: {
    fontSize: 12,
    color: '#FF9800',
    marginRight: 12,
  },
  hospitalOpenStatus: {
    fontSize: 12,
    fontWeight: '500',
  },
  hospitalActions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    marginTop: 8,
  },
  actionButton: {
    backgroundColor: '#0A2F52',
    paddingVertical: 6,
    paddingHorizontal: 12,
    borderRadius: 4,
    marginLeft: 8,
  },
  actionButtonText: {
    color: 'white',
    fontSize: 12,
    fontWeight: '600',
  },
});

export default NearbyHospitals;
