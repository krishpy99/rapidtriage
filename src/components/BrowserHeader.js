import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

const BrowserHeader = ({ url = 'rapidtriage.ai' }) => {
  return (
    <View style={styles.header}>
      <View style={styles.addressBar}>
        <Text style={styles.addressText}>{url}</Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  header: {
    backgroundColor: '#E5E5E5',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#DDDDDD',
  },
  addressBar: {
    backgroundColor: '#FFFFFF',
    borderRadius: 20,
    paddingVertical: 6,
    paddingHorizontal: 12,
    alignItems: 'center',
  },
  addressText: {
    color: '#333333',
    fontSize: 14,
  },
});

export default BrowserHeader;
