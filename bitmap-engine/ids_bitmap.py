#!/usr/bin/env python
#coding: utf-8

import json
import argparse
import struct


def get_ids(filename):
    LEN = 4
    ids = []
    with open(filename, 'rb') as f:
        data = f.read()
        for i in range(len(data)/LEN):
            value = struct.unpack('I', data[i*LEN:(i+1)*LEN])[0]
            ids.append(value)
    return ids

def get_bitmap_blocks(filename):
    LEN = 8
    blocks = []
    with open(filename, 'rb') as f:
        data = f.read()
        for i in range(len(data)/LEN):
            part = data[i*LEN:(i+1)*LEN]
            value = struct.unpack('Q', part)[0]
            blocks.append((value, part))
    return blocks

# ==============================================================================
# functions
# ==============================================================================
def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-v', '--verbose', action='store_true', help='Print more info')
    sub_parsers = parser.add_subparsers(help='', dest='cmd')

    make_parser = sub_parsers.add_parser('make')
    make_parser.add_argument('-f', '--filename', metavar='FILE', required=True, help='Target file name')
    make_parser.add_argument('-j', '--json', metavar='FILE', help='JSON ids file')
    make_parser.add_argument('-d', '--data', metavar='STRING', help='Raw ids number')
    make_parser.add_argument('-s', '--start', metavar='INT', type=int, help='Start number (include, {start}>0)')
    make_parser.add_argument('-e', '--end', metavar='INT', type=int, help='End number (not include)')

    show_parser = sub_parsers.add_parser('show')
    show_parser.add_argument('-i', '--ids', metavar='FILE', help='Ids file')
    show_parser.add_argument('-b', '--bitmap', metavar='FILE', help='Bitmap file')
    args = parser.parse_args()
    print 'Args:', args
    return args


def make_ids(args):
    if args.json:
        with open(args.json, 'r') as f:
            ids = json.load(f)
    elif args.data:
        ids = [int(n) for n in args.data.split()]
    else:
        ids = range(args.start, args.end)

    with open(args.filename, 'wb') as f:
        for i in ids:
            value = struct.pack('I', i)
            if args.verbose: print i
            f.write(value)
    print 'Written to: [{}]'.format(args.filename)


def show(args):
    if args.ids:
        print 'ID list:'
        print '--------'
        ids = get_ids(args.ids)
        cnt = 0
        for uid in ids:
            if args.verbose: print '%d,' % uid,
            cnt += 1
            if cnt % 10 == 0:
                if args.verbose: print ''
        print ' > DONE\n'

    if args.bitmap:
        print 'Bitmap blocks:'
        print '--------------'
        blocks = get_bitmap_blocks(args.bitmap)
        for value, part in blocks:
            print '{:064b}, {}'.format(value, ' '.join(['%02X' % ord(c) for c in part]))

def main():
    args = parse_args()
    {'make': make_ids, 'show': show}[args.cmd](args)

if __name__ == '__main__':
    main()
