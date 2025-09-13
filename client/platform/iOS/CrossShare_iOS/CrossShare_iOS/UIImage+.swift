//
//  UIImage+.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/7.
//

import Foundation
import UIKit
import AVKit
import AVFoundation

extension UIImage {

    var cmSampleBuffer: CMSampleBuffer? {
        guard let jpegData = jpegData(compressionQuality: 1) else { return nil }
        return sampleBufferFromJPEGData(jpegData)
    }

    private func sampleBufferFromJPEGData(_ jpegData: Data) -> CMSampleBuffer? {

        guard let cgImage = cgImage else { return nil }
        let rawPixelSize = CGSize(width: cgImage.width, height: cgImage.height)

        var format: CMFormatDescription? = nil
        let _ = CMVideoFormatDescriptionCreate(
            allocator: kCFAllocatorDefault,
            codecType: kCMVideoCodecType_JPEG,
            width: Int32(rawPixelSize.width),
            height: Int32(rawPixelSize.height),
            extensions: nil,
            formatDescriptionOut: &format)

        do {
            let cmBlockBuffer = try jpegData.toCMBlockBuffer()
            var size = jpegData.count
            var sampleBuffer: CMSampleBuffer? = nil
            let nowTime = CMTime(seconds: CACurrentMediaTime(), preferredTimescale: 60)
            let _1_60_s = CMTime(value: 1, timescale: 60)

            var timingInfo: CMSampleTimingInfo = CMSampleTimingInfo(
                duration: _1_60_s,
                presentationTimeStamp: nowTime,
                decodeTimeStamp: .invalid)

            let _ = CMSampleBufferCreateReady(
                allocator: kCFAllocatorDefault,
                dataBuffer: cmBlockBuffer,
                formatDescription: format,
                sampleCount: 1,
                sampleTimingEntryCount: 1,
                sampleTimingArray: &timingInfo,
                sampleSizeEntryCount: 1,
                sampleSizeArray: &size,
                sampleBufferOut: &sampleBuffer)

            if sampleBuffer != nil {
                return sampleBuffer

            } else {
                Logger.info("sampleBuffer is nil")
                return nil
            }
        } catch {
            Logger.info("error ugh \(error.localizedDescription)")
            return nil
        }
    }

    func pixelBuffer() -> CVPixelBuffer? {
        let width = Int(size.width)
        let height = Int(size.height)

        let attributes: [String: Any] = [
            kCVPixelBufferCGImageCompatibilityKey as String: true,
            kCVPixelBufferCGBitmapContextCompatibilityKey as String: true
        ]

        var pixelBuffer: CVPixelBuffer?
        let status = CVPixelBufferCreate(
            kCFAllocatorDefault,
            width,
            height,
            kCVPixelFormatType_32ARGB,
            attributes as CFDictionary,
            &pixelBuffer
        )

        guard status == kCVReturnSuccess, let buffer = pixelBuffer else {
            return nil
        }

        CVPixelBufferLockBaseAddress(buffer, CVPixelBufferLockFlags(rawValue: 0))
        defer { CVPixelBufferUnlockBaseAddress(buffer, CVPixelBufferLockFlags(rawValue: 0)) }

        let context = CGContext(
            data: CVPixelBufferGetBaseAddress(buffer),
            width: width,
            height: height,
            bitsPerComponent: 8,
            bytesPerRow: CVPixelBufferGetBytesPerRow(buffer),
            space: CGColorSpaceCreateDeviceRGB(),
            bitmapInfo: CGImageAlphaInfo.noneSkipFirst.rawValue
        )

        guard let cgContext = context, let cgImage = self.cgImage else {
            return nil
        }

        cgContext.draw(cgImage, in: CGRect(x: 0, y: 0, width: width, height: height))

        return buffer
    }
}

