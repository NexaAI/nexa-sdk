
schema = {
  "type": "object",
  "properties": {
    "victim_info": {
      "type": "object",
      "properties": {
        "id": {"type": "string"},
        "emergency_status": {
                "type": "string",
                "enum": ["critical", "very_urgent", "urgent", "stable", "unknown"]
                },       
        "location": {
          "type": "object",
          "properties": {
            "lat": {"type": "number"},
            "lon": {"type": "number"},
            "details": {"type": "string"},
            "nearest_landmark": {"type": "string"}
          },
        
        },
        "personal_info": {
          "type": "object",
          "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer"},
            "gender": {"type": "string"},
            "language": {"type": "string"},
            "physical_description": {"type": "string"}
          },
        
        },
        "medical_info": {
          "type": "object",
          "properties": {
            "injuries": {"type": "array", "items": {"type": "string"}},
            "pain_level": {"type": "integer"},
            "medical_conditions": {"type": "array", "items": {"type": "string"}},
            "medications": {"type": "array", "items": {"type": "string"}},
            "allergies": {"type": "array", "items": {"type": "string"}},
            "blood_type": {"type": "string"}
          },
        
        },
        "situation": {
          "type": "object",
          "properties": {
            "disaster_type": {"type": "string"},
            "immediate_needs": {"type": "array", "items": {"type": "string"}},
            "trapped": {"type": "boolean"},
            "mobility": {"type": "string"},
            "nearby_hazards": {"type": "array", "items": {"type": "string"}}
          },
        
        },
        "contact_info": {
          "type": "object",
          "properties": {
            "phone": {"type": "string"},
            "email": {"type": "string"},
            "emergency_contact": {
              "type": "object",
              "properties": {
                "name": {"type": "string"},
                "relationship": {"type": "string"},
                "phone": {"type": "string"}
              },
            
            }
          },
        
        },
        "resources": {
          "type": "object",
          "properties": {
            "food_status": {"type": "string"},
            "water_status": {"type": "string"},
            "shelter_status": {"type": "string"},
            "communication_devices": {"type": "array", "items": {"type": "string"}}
          },
        
        },
        "rescue_info": {
          "type": "object",
          "properties": {
            "last_contact": {"type": "string"},
            "rescue_team_eta": {"type": "string"},
            "special_rescue_needs": {"type": "string"}
          },
        
        },
        "environmental_data": {
          "type": "object",
          "properties": {
            "temperature": {"type": "number"},
            "humidity": {"type": "number"},
            "air_quality": {"type": "string"},
            "weather": {"type": "string"}
          },
        
        },
        "device_data": {
          "type": "object",
          "properties": {
            "battery_level": {"type": "integer"},
            "network_status": {"type": "string"}
          },
        
        },
        "social_info": {
          "type": "object",
          "properties": {
            "group_size": {"type": "integer"},
            "dependents": {"type": "integer"},
            "nearby_victims_count": {"type": "integer"},
            "can_communicate_verbally": {"type": "boolean"}
          },
        
        },
        "psychological_status": {
          "type": "object",
          "properties": {
            "stress_level": {"type": "string"},
            "special_needs": {"type": "string"}
          },
        
        }
      },
    
    }
  },

}

