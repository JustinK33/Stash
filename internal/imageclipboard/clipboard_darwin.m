#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

static NSData *stash_png_data(NSData *sourceData) {
    NSBitmapImageRep *representation = [NSBitmapImageRep imageRepWithData:sourceData];
    if (representation == nil) {
        return nil;
    }
    return [representation representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
}

int stash_read_clipboard_png(unsigned char **outputData, long *outputLength) {
    @autoreleasepool {
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSData *pngData = [pasteboard dataForType:NSPasteboardTypePNG];

        if (pngData == nil) {
            NSData *tiffData = [pasteboard dataForType:NSPasteboardTypeTIFF];
            if (tiffData != nil) {
                pngData = stash_png_data(tiffData);
            }
        }

        if (pngData == nil) {
            NSArray<NSURL *> *urls = [pasteboard readObjectsForClasses:@[[NSURL class]]
                                                               options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];
            NSURL *url = urls.firstObject;
            if (url != nil) {
                NSData *fileData = [NSData dataWithContentsOfURL:url];
                if (fileData != nil) {
                    pngData = stash_png_data(fileData);
                }
            }
        }

        if (pngData == nil || pngData.length == 0) {
            return 0;
        }

        *outputLength = (long)pngData.length;
        *outputData = malloc((size_t)*outputLength);
        if (*outputData == NULL) {
            return 0;
        }
        memcpy(*outputData, pngData.bytes, (size_t)*outputLength);
        return 1;
    }
}

int stash_write_clipboard_png(const unsigned char *data, long length) {
    @autoreleasepool {
        NSData *pngData = [NSData dataWithBytes:data length:(NSUInteger)length];
        if (stash_png_data(pngData) == nil) {
            return 0;
        }

        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        return [pasteboard setData:pngData forType:NSPasteboardTypePNG] ? 1 : 0;
    }
}
