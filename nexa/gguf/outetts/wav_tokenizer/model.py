import torch
from torch import nn
import os
import json
import torch

class WavEncoder(nn.Module):
    """Handles only the encoding part of WavTokenizer"""
    
    def __init__(self, feature_extractor):
        super().__init__()
        self.feature_extractor = feature_extractor
        
    def _create_config(self):
        config = {
            'feature_extractor_class': 'decoder.feature_extractors.EncodecFeatures',
            'feature_extractor_config': {
                'encodec_model': 'encodec_24khz',
                'bandwidths': [6.6, 6.6, 6.6, 6.6],
                'train_codebooks': True,
                'num_quantizers': 1,
                'dowmsamples': [8, 5, 4, 2],
                'vq_bins': 4096,
                'vq_kmeans': 200
            }
        }
        return config
    
    def save_pretrained(self, save_directory):
        """Save the model and its configuration to a directory"""
        os.makedirs(save_directory, exist_ok=True)
        
        model_path = os.path.join(save_directory, "encoder_model.pt")
        torch.save(self.state_dict(), model_path)
        
        config = self._create_config()
        config_path = os.path.join(save_directory, "config.json")
        with open(config_path, 'w') as f:
            json.dump(config, f, indent=2)
    
    @classmethod
    def from_pretrained(cls, model_directory):
        """Load a model and its configuration from a directory"""
        config_path = os.path.join(model_directory, "config.json")
        with open(config_path, 'r') as f:
            config = json.load(f)
        
        # Import feature extractor class
        module_name, class_name = config['feature_extractor_class'].rsplit('.', 1)
        module = __import__(module_name, fromlist=[class_name])
        feature_extractor_cls = getattr(module, class_name)
        
        # Create feature extractor
        feature_extractor = feature_extractor_cls(**config['feature_extractor_config'])
        
        model = cls(feature_extractor)
        
        model_path = os.path.join(model_directory, "encoder_model.pt")
        state_dict = torch.load(model_path, map_location='cpu')
        model.load_state_dict(state_dict)
        
        return model

    @torch.inference_mode()
    def forward(self, audio_input, **kwargs):
        """Encode audio into discrete codes"""
        features, discrete_codes, _ = self.feature_extractor.infer(audio_input, **kwargs)
        return features, discrete_codes

class WavDecoder(nn.Module):
    """Handles only the decoding part of WavTokenizer"""
    
    def __init__(self, backbone, head, codebook_weights):
        super().__init__()
        self.backbone = backbone
        self.head = head
        self.register_buffer('codebook_weights', codebook_weights)
    
    def _create_config(self):
        config = {
            'backbone_class': 'decoder.models.VocosBackbone',
            'head_class': 'decoder.heads.ISTFTHead',
            'backbone_config': {
                'input_channels': 512,
                'dim': 768,
                'intermediate_dim': 2304,
                'num_layers': 12,
                'adanorm_num_embeddings': 4
            },
            'head_config': {
                'dim': 768,
                'n_fft': 1280,
                'hop_length': 320,
                'padding': 'same'
            }
        }
        return config
    
    def save_pretrained(self, save_directory):
        """Save the model and its configuration to a directory"""
        os.makedirs(save_directory, exist_ok=True)
        
        model_path = os.path.join(save_directory, "decoder_model.pt")
        save_dict = {
            'model_state_dict': self.state_dict(),
            'codebook_weights': self.codebook_weights
        }
        torch.save(save_dict, model_path)
        
        config = self._create_config()
        config_path = os.path.join(save_directory, "config.json")
        with open(config_path, 'w') as f:
            json.dump(config, f, indent=2)
    
    @classmethod
    def from_pretrained(cls, model_directory):
        """Load a model and its configuration from a directory"""
        config_path = os.path.join(model_directory, "config.json")
        with open(config_path, 'r') as f:
            config = json.load(f)
        
        # Import and create backbone
        module_name, class_name = config['backbone_class'].rsplit('.', 1)
        module = __import__(module_name, fromlist=[class_name])
        backbone_cls = getattr(module, class_name)
        backbone = backbone_cls(**config['backbone_config'])
        
        # Import and create head
        module_name, class_name = config['head_class'].rsplit('.', 1)
        module = __import__(module_name, fromlist=[class_name])
        head_cls = getattr(module, class_name)
        head = head_cls(**config['head_config'])
       
        model_path = os.path.join(model_directory, "decoder_model.pt")
        checkpoint = torch.load(model_path, map_location='cpu')
        
        model = cls(backbone, head, checkpoint['codebook_weights'])
        model.load_state_dict(checkpoint['model_state_dict'])
        
        return model

    def codes_to_features(self, codes):
        """Convert discrete codes to features using codebook"""
        if codes.dim() == 2:
            codes = codes.unsqueeze(1)
            
        n_bins = self.codebook_weights.size(0) // len(codes)
        offsets = torch.arange(0, n_bins * len(codes), n_bins, device=codes.device)
        embeddings_idxs = codes + offsets.view(-1, 1, 1)
        
        features = torch.nn.functional.embedding(embeddings_idxs, self.codebook_weights).sum(dim=0)
        features = features.transpose(1, 2)
        
        return features

    @torch.inference_mode()
    def forward(self, features_input, **kwargs):
        """Decode features to audio"""
        x = self.backbone(features_input, **kwargs)
        audio_output = self.head(x)
        return audio_output