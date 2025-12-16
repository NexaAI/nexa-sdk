import Foundation
import SwiftUI

class ImageCache {
    static let shared = NSCache<NSString, UIImage>()

    static func loadImage(at url: URL) -> UIImage? {
        let key = url.path() as NSString
        if let cached = ImageCache.shared.object(forKey: key) {
            return cached
        }
        if let image = UIImage(contentsOfFile: url.path()) {
            ImageCache.shared.setObject(image, forKey: key)
            return image
        }
        return nil
    }
    
    static func loadImage(at url: URL?) async -> UIImage? {
        guard let url else { return nil }
        let key = url.path() as NSString
        if let cached = ImageCache.shared.object(forKey: key) {
            return cached
        }

        let image = await withCheckedContinuation { cont in
            DispatchQueue.global().async {
                let img = thumbnailImage(of: url)
                cont.resume(returning: img)
            }
        }

        if let image {
            ImageCache.shared.setObject(image, forKey: key)
        }
        return image
    }

    private static func thumbnailImage(of url: URL?) -> UIImage? {
        guard let url else {
            return nil
        }
        guard let imageSource = CGImageSourceCreateWithURL(url as CFURL, nil) else {
            return nil
        }

        let options: [NSString: Any] = [
            kCGImageSourceThumbnailMaxPixelSize: 600,
            kCGImageSourceCreateThumbnailFromImageAlways: true
        ]

        if let cgImage = CGImageSourceCreateThumbnailAtIndex(imageSource, 0, options as CFDictionary) {
            return UIImage(cgImage: cgImage)
        }
        return nil
    }
}

struct LocalAsyncImage<Content: View>: View {
    let url: URL?
    let content: (Image) -> Content

    @State private var uiImage: UIImage?

    var body: some View {
        ZStack {
            if let uiImage {
                content(Image(uiImage: uiImage))
            }
        }
        .task {
            if uiImage == nil {
                uiImage = await ImageCache.loadImage(at: url)
            }
        }
    }
}
